package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestRecoveryIssueAndValidate(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	user, err := m.ValidateRecovery(ctx, "alice", secret)
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
}

func TestRecoveryWrongSecret(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	if _, err := m.IssueRecovery(ctx, "alice"); err != nil {
		require.NoError(t, err)
	}

	_, err = m.ValidateRecovery(ctx, "alice", "wrong-secret")
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 복구 비밀이 없는 계정은 링크를 만들어 낼 수 없다
func TestRecoveryWithoutIssue(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	_, err = m.ValidateRecovery(ctx, "alice", "anything")
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

func TestRecoveryRevoke(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	secret, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	require.NoError(t, m.RevokeRecovery(ctx, "alice"))

	// 한 번 쓴 링크는 다시 통하지 않는다
	_, err = m.ValidateRecovery(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)

	// 없는 것을 지워도 에러가 아니다
	assert.NoError(t, m.RevokeRecovery(ctx, "alice"))
}

func TestRecoveryResendInterval(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	_, err = m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	_, err = m.IssueRecovery(ctx, "alice")
	assert.ErrorIs(t, err, auth.ErrRecoveryTooSoon)
}

// 재발송 간격이 지나면 새 비밀이 나오고 옛것은 죽는다
func TestRecoveryReissueAfterInterval(t *testing.T) {
	ctx := t.Context()
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	old, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	// 발급 시각을 간격 너머로 되돌린다
	require.NoError(t, db.Model(&auth.Recovery{}).
		Where("username = ?", "alice").
		Update("created_at", time.Now().Add(-auth.RecoveryResendInterval-time.Minute)).Error)

	fresh, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)
	assert.NotEqual(t, old, fresh)

	_, err = m.ValidateRecovery(ctx, "alice", old)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)

	user, err := m.ValidateRecovery(ctx, "alice", fresh)
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
}

func TestRecoveryExpired(t *testing.T) {
	ctx := t.Context()
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	secret, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	require.NoError(t, db.Model(&auth.Recovery{}).
		Where("username = ?", "alice").
		Update("expired_at", time.Now().Add(-time.Minute)).Error)

	_, err = m.ValidateRecovery(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 복구 비밀은 계정마다 따로다
func TestRecoveryIsolatedPerUser(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)
	_, _, err = m.Register(ctx, "bob", "password")
	require.NoError(t, err)

	secret, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	_, err = m.ValidateRecovery(ctx, "bob", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 복구 비밀은 로그인 자격증명 목록에 섞이지 않는다
func TestRecoveryNotInTokenList(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	_, err = m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	tokens, err := m.ListToken(ctx, "alice")
	require.NoError(t, err)

	for _, token := range tokens {
		assert.NotEqual(t, auth.TokenKind("recovery"), token.Kind)
	}
	assert.Len(t, tokens, 1) // password 하나뿐이다
}

func TestFindUserByEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)
	_, _, err = m.Register(ctx, "bob", "password")
	require.NoError(t, err)

	user, err := m.FindUserByEmail(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)

	_, err = m.FindUserByEmail(ctx, "nobody@example.com")
	assert.Error(t, err)

	// 이메일이 없는 계정이 빈 문자열로 걸려 나오면 안 된다
	_, err = m.FindUserByEmail(ctx, "")
	assert.Error(t, err)
}

func TestUpdateEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithGroup("user"))
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@example.com"))

	user, err := m.GetUser(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	// 이메일만 바뀌고 나머지는 그대로다
	require.Len(t, user.Groups, 1)
	assert.Equal(t, "user", user.Groups[0].Name)

	found, err := m.FindUserByEmail(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "alice", found.Name)

	assert.Error(t, m.UpdateEmail(ctx, "nobody", "nobody@example.com"))
}
