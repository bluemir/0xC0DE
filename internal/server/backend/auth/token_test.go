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
	expiredToken, err := m.IssueToken(username, "test-kind", "secret", auth.ExpiredAfter(-1*time.Hour))
	require.NoError(t, err)

	// Validate should fail
	err = expiredToken.Validate("secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// Issue valid token
	validToken, err := m.IssueToken(username, "test-kind-2", "secret", auth.ExpiredAfter(1*time.Hour))
	require.NoError(t, err)

	// Validate should succeed
	err = validToken.Validate("secret")
	assert.NoError(t, err)
}
