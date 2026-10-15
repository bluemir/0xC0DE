package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestVerificationIssueAndConfirm(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	user, _, err := m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)
	assert.Nil(t, user.EmailVerifiedAt)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	confirmed, err := m.ConfirmVerification(ctx, "alice", secret)
	require.NoError(t, err)
	require.NotNil(t, confirmed.EmailVerifiedAt)
	assert.WithinDuration(t, time.Now(), *confirmed.EmailVerifiedAt, time.Minute)

	// 저장도 됐는지 다시 읽어 확인한다
	stored, err := m.GetUser(ctx, "alice")
	require.NoError(t, err)
	require.NotNil(t, stored.EmailVerifiedAt)
}

func TestVerificationWrongSecret(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	_, err = m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", "wrong-secret")
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 발급한 적이 없으면 링크를 만들어 낼 수 없다
func TestVerificationWithoutIssue(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", "anything")
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

func TestVerificationEmptyEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password")
	require.NoError(t, err)

	_, err = m.IssueVerification(ctx, "alice", "")
	assert.Error(t, err)
}

// 한 번 쓴 링크는 다시 쓰이지 않는다
func TestVerificationConfirmTwice(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

func TestVerificationExpired(t *testing.T) {
	ctx := t.Context()
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	// 만료 시각을 지나간 것으로 되돌린다
	require.NoError(t, db.Model(&auth.Verification{}).
		Where("username = ?", "alice").
		Update("expired_at", time.Now().Add(-time.Minute)).Error)

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 주소를 바꾸면 이전 주소로 보낸 링크는 폐기된다
func TestVerificationEmailChangedAfterIssue(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@other.example.com"))

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

func TestVerificationResendInterval(t *testing.T) {
	ctx := t.Context()
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	_, err = m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	_, err = m.IssueVerification(ctx, "alice", "alice@example.com")
	assert.ErrorIs(t, err, auth.ErrVerificationTooSoon)

	// 간격이 지나면 다시 보낼 수 있다
	require.NoError(t, db.Model(&auth.Verification{}).
		Where("username = ?", "alice").
		Update("created_at", time.Now().Add(-auth.VerificationResendInterval-time.Minute)).Error)

	_, err = m.IssueVerification(ctx, "alice", "alice@example.com")
	assert.NoError(t, err)
}

// 확인 메일은 백그라운드로 나가므로, 주소를 두 번 빠르게 바꾸면 먼저 걸린 발송이
// 뒤늦게 이전 주소로 행을 남길 수 있다. 그 링크는 지금 주소를 증명하지 못한다.
func TestVerificationForStaleEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@other.example.com"))

	// 이전 주소를 들고 뒤늦게 돈 발송
	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	assert.ErrorIs(t, err, auth.ErrUnauthorized)

	user, err := m.GetUser(ctx, "alice")
	require.NoError(t, err)
	assert.Nil(t, user.EmailVerifiedAt)
}

// 주소가 달라지면 간격과 상관없이 새로 발급한다
func TestVerificationResendForOtherEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	_, err = m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@other.example.com"))

	secret, err := m.IssueVerification(ctx, "alice", "alice@other.example.com")
	require.NoError(t, err)

	_, err = m.ConfirmVerification(ctx, "alice", secret)
	assert.NoError(t, err)
}

// 주소를 바꾸면 확인 상태가 풀린다
func TestUpdateEmailResetsVerified(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)
	_, err = m.ConfirmVerification(ctx, "alice", secret)
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@other.example.com"))

	user, err := m.GetUser(ctx, "alice")
	require.NoError(t, err)
	assert.Equal(t, "alice@other.example.com", user.Email)
	assert.Nil(t, user.EmailVerifiedAt)
}

// 같은 주소를 다시 저장하는 것은 확인 상태를 건드리지 않는다
func TestUpdateEmailSameAddressKeepsVerified(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	secret, err := m.IssueVerification(ctx, "alice", "alice@example.com")
	require.NoError(t, err)
	_, err = m.ConfirmVerification(ctx, "alice", secret)
	require.NoError(t, err)

	require.NoError(t, m.UpdateEmail(ctx, "alice", "alice@example.com"))

	user, err := m.GetUser(ctx, "alice")
	require.NoError(t, err)
	assert.NotNil(t, user.EmailVerifiedAt)
}

// 확인하지 않은 주소로도 복구는 된다
func TestRecoveryWorksWithUnverifiedEmail(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	user, err := m.FindUserByEmail(ctx, "alice@example.com")
	require.NoError(t, err)
	require.Nil(t, user.EmailVerifiedAt)

	secret, err := m.IssueRecovery(ctx, "alice")
	require.NoError(t, err)

	_, err = m.ValidateRecovery(ctx, "alice", secret)
	assert.NoError(t, err)
}
