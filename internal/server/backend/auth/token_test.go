package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestTokenManagement(t *testing.T) {
	m := newTestManager(t)
	username := "user-token-test"
	_, err := m.CreateUser(username)
	require.NoError(t, err)

	// Generate Access Key
	token, key, err := m.GenerateAccessKey(username)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.NotEmpty(t, key)
	assert.Equal(t, auth.TokenKindAccessKey, token.Kind)

	// Validate Access Key (key format: username.index.secret)
	// The Validate method takes the "unhashedSecret" which is the last part of key.
	// But `GenerateAccessKey` returns the full key string.
	// However, `token.Validate` expects the raw secret.
	// The `key` returned by `GenerateAccessKey` is composite.
	// Usage in system likely parses the key string to find username/index, then looks up token, then validates secret.
	// But here we test `token.Validate` logic or `auth.Manager` logic?
	// `auth.Manager` doesn't have a `ValidateKey` method shown in interface in auth.go (only Validate on Token struct).
	// But `Default` method uses `GetToken` then `Validate`.
	// Let's verify we can find and validate manually.

	// List Tokens
	tokens, err := m.ListToken(username)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, token.Index, tokens[0].Index)

	// Revoke Token
	err = m.RevokeToken(username, auth.TokenKindAccessKey, token.Index)
	assert.NoError(t, err)

	// List again
	tokens, err = m.ListToken(username)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestTokenExpiration(t *testing.T) {
	m := newTestManager(t)
	username := "user-expire-test"
	_, err := m.CreateUser(username)
	require.NoError(t, err)

	// Issue expired token
	expiredToken, err := m.IssueToken(username, auth.TokenKindAccessKey, "secret", auth.ExpiredAfter(-1*time.Hour))
	require.NoError(t, err)

	// Validate should fail
	err = expiredToken.Validate("secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// Issue valid token
	validToken, err := m.IssueToken(username, auth.TokenKindAccessKey, "secret", auth.ExpiredAfter(1*time.Hour))
	require.NoError(t, err)

	// Validate should succeed
	err = validToken.Validate("secret")
	assert.NoError(t, err)
}

func TestTokenSecret(t *testing.T) {
	m := newTestManager(t)
	username := "user-secret-test"
	_, err := m.CreateUser(username)
	require.NoError(t, err)

	// 해당하는 것만 채워지고 나머지는 nil 로 남아야 한다
	issued, err := m.IssueToken(username, auth.TokenKindPassword, "secret")
	require.NoError(t, err)

	got, err := m.GetToken(username, auth.TokenKindPassword, issued.Index)
	require.NoError(t, err)
	require.NotNil(t, got.Secret.Password)
	assert.Nil(t, got.Secret.AccessKey)
	assert.Nil(t, got.Secret.Passkey)
	assert.Empty(t, got.CredentialID)
	assert.NoError(t, got.Validate("secret"))

	key, _, err := m.GenerateAccessKey(username)
	require.NoError(t, err)

	gotKey, err := m.GetToken(username, auth.TokenKindAccessKey, key.Index)
	require.NoError(t, err)
	require.NotNil(t, gotKey.Secret.AccessKey)
	assert.Nil(t, gotKey.Secret.Password)

	// UpdatePassword 도 Secret 을 통째로 갈아끼운다
	require.NoError(t, m.UpdatePassword(username, "next"))

	updated, err := m.GetToken(username, auth.TokenKindPassword, 0)
	require.NoError(t, err)
	require.NotNil(t, updated.Secret.Password)
	assert.NoError(t, updated.Validate("next"))
	assert.Error(t, updated.Validate("secret"))

	// passkey 는 나눠 가진 비밀이 없으므로 Validate 로 통과할 수 없다
	passkey := auth.Token{
		Username:     username,
		Kind:         auth.TokenKindPasskey,
		CredentialID: []byte("credential-id"),
		Secret:       auth.Secret{Passkey: &auth.PasskeySecret{Credential: []byte(`{}`)}},
	}
	assert.Error(t, passkey.Validate(""))
}

func TestTokenSecretMismatch(t *testing.T) {
	m := newTestManager(t)
	username := "user-mismatch-test"
	_, err := m.CreateUser(username)
	require.NoError(t, err)

	// Kind 와 Secret 이 어긋나면 저장되지 않아야 한다
	_, err = m.IssueToken(username, auth.TokenKindPasskey, "secret")
	assert.Error(t, err)

	_, err = m.IssueToken(username, "no-such-kind", "secret")
	assert.Error(t, err)

	tokens, err := m.ListToken(username)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}
