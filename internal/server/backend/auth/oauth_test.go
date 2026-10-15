package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestOAuthLink(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	_, err := m.CreateUser(ctx, "oauth-user")
	require.NoError(t, err)

	token, err := m.LinkOAuth(ctx, "oauth-user", auth.TokenKindGoogle, "google-subject-1", auth.OAuthSecret{
		Email: "user@example.com",
		Name:  "Example User",
	})
	require.NoError(t, err)
	require.NotNil(t, token.Secret.OAuth)
	assert.Nil(t, token.Secret.Password)
	assert.Equal(t, []byte("google-subject-1"), token.ExternalID)
	assert.False(t, token.Secret.OAuth.RefreshedAt.IsZero())

	// provider 가 준 ID 로 사용자를 역추적한다
	found, err := m.FindOAuthUser(ctx, auth.TokenKindGoogle, "google-subject-1")
	require.NoError(t, err)
	assert.Equal(t, "oauth-user", found.Name)

	// 같은 값이라도 provider 가 다르면 별개다
	_, err = m.FindOAuthUser(ctx, auth.TokenKindGitHub, "google-subject-1")
	assert.Error(t, err)

	_, err = m.LinkOAuth(ctx, "oauth-user", auth.TokenKindGitHub, "google-subject-1", auth.OAuthSecret{})
	require.NoError(t, err)

	links, err := m.ListOAuth(ctx, "oauth-user")
	require.NoError(t, err)
	assert.Len(t, links, 2)

	// 연결을 끊으면 역추적되지 않는다
	require.NoError(t, m.RevokeOAuth(ctx, "oauth-user", auth.TokenKindGoogle, token.Index))

	_, err = m.FindOAuthUser(ctx, auth.TokenKindGoogle, "google-subject-1")
	assert.Error(t, err)

	links, err = m.ListOAuth(ctx, "oauth-user")
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestOAuthLinkIsExclusive(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	for _, name := range []string{"first", "second"} {
		_, err := m.CreateUser(ctx, name)
		require.NoError(t, err)
	}

	_, err := m.LinkOAuth(ctx, "first", auth.TokenKindGoogle, "shared-subject", auth.OAuthSecret{})
	require.NoError(t, err)

	// 같은 외부 계정이 두 사용자에 붙으면 안 된다
	_, err = m.LinkOAuth(ctx, "second", auth.TokenKindGoogle, "shared-subject", auth.OAuthSecret{})
	assert.Error(t, err)
}

func TestOAuthProfileUpdate(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	_, err := m.CreateUser(ctx, "profile-user")
	require.NoError(t, err)

	past := time.Now().Add(-24 * time.Hour)
	_, err = m.LinkOAuth(ctx, "profile-user", auth.TokenKindGitHub, "gh-1", auth.OAuthSecret{
		Email:       "old@example.com",
		RefreshedAt: past,
	})
	require.NoError(t, err)

	require.NoError(t, m.UpdateOAuthProfile(ctx, auth.TokenKindGitHub, "gh-1", auth.OAuthSecret{
		Email: "new@example.com",
		Name:  "New Name",
	}))

	links, err := m.ListOAuth(ctx, "profile-user")
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.NotNil(t, links[0].Secret.OAuth)
	assert.Equal(t, "new@example.com", links[0].Secret.OAuth.Email)
	assert.Equal(t, "New Name", links[0].Secret.OAuth.Name)
	assert.True(t, links[0].Secret.OAuth.RefreshedAt.After(past))

	// 없는 연결은 error 다
	assert.Error(t, m.UpdateOAuthProfile(ctx, auth.TokenKindGitHub, "no-such", auth.OAuthSecret{}))
}

func TestOAuthRejectsNonProvider(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)
	_, err := m.CreateUser(ctx, "reject-user")
	require.NoError(t, err)

	// passkey 나 password 는 이 경로로 다루지 않는다
	_, err = m.LinkOAuth(ctx, "reject-user", auth.TokenKindPasskey, "x", auth.OAuthSecret{})
	assert.Error(t, err)

	_, err = m.FindOAuthUser(ctx, auth.TokenKindPassword, "x")
	assert.Error(t, err)

	assert.Error(t, m.RevokeOAuth(ctx, "reject-user", auth.TokenKindPassword, 0))

	// ID 가 비면 역추적할 수 없다
	_, err = m.LinkOAuth(ctx, "reject-user", auth.TokenKindGoogle, "", auth.OAuthSecret{})
	assert.Error(t, err)
}
