package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func newTestManager(t *testing.T) *auth.Manager {
	ctx := t.Context()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	events, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	m, err := auth.New(ctx, &auth.Config{Salt: "test-salt"}, db, events)
	require.NoError(t, err)
	return m
}

func TestRegisterAndLogin(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)

	// Register
	user, token, err := m.Register(ctx, "testuser", "password", auth.WithGroup("user"))
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, token)
	assert.Equal(t, "testuser", user.Name)
	assert.Equal(t, "testuser", user.Name)
	groupNames := []string{}
	for _, g := range user.Groups {
		groupNames = append(groupNames, g.Name)
	}
	assert.Contains(t, groupNames, "user")

	// Login (Default)
	loggedInUser, err := m.Default(ctx, "testuser", "password")
	require.NoError(t, err)
	require.NotNil(t, loggedInUser)
	assert.Equal(t, user.Name, loggedInUser.Name)

	// Login fail
	_, err = m.Default(ctx, "testuser", "wrongpassword")
	assert.Error(t, err)
}

func TestTokenLifecycle(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	m.Register(ctx, "user1", "pass")

	// Issue Token
	token, err := m.IssueToken(ctx, "user1", auth.TokenKindAccessKey, "secret123", auth.ExpiredAfter(1*time.Hour))
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, 0, token.Index)

	// Validate Token
	err = token.Validate("secret123")
	assert.NoError(t, err)

	err = token.Validate("wrong")
	assert.Error(t, err)

	// Issue another token (index increment)
	token2, err := m.IssueToken(ctx, "user1", auth.TokenKindAccessKey, "secret456")
	assert.NoError(t, err)
	assert.Equal(t, 1, token2.Index)

	// List Tokens
	tokens, err := m.ListToken(ctx, "user1")
	assert.NoError(t, err)
	// Expect 3 tokens? Password token (index 0) + 2 access keys?
	// Register creates a password token.
	// IssueToken(AccessKey) creates access key tokens.
	// But ListToken gets ALL tokens?
	// Let's check ListToken implementation. It filters by Username only.
	// So 1 (password) + 2 (access keys) = 3 total.
	assert.Len(t, tokens, 3)

	// Get Token
	gotToken, err := m.GetToken(ctx, "user1", auth.TokenKindAccessKey, 1)
	assert.NoError(t, err)
	assert.Equal(t, token2.Secret.AccessKey.HashedSecret, gotToken.Secret.AccessKey.HashedSecret)

	// Revoke Token
	err = m.RevokeToken(ctx, "user1", auth.TokenKindAccessKey, 0)
	assert.NoError(t, err)

	_, err = m.GetToken(ctx, "user1", auth.TokenKindAccessKey, 0)
	assert.Error(t, err)
}

func TestUpdatePassword(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	m.Register(ctx, "user1", "oldpass")

	err := m.UpdatePassword(ctx, "user1", "newpass")
	assert.NoError(t, err)

	_, err = m.Default(ctx, "user1", "newpass")
	assert.NoError(t, err)

	_, err = m.Default(ctx, "user1", "oldpass")
	assert.Error(t, err)
}
