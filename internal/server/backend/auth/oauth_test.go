package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestOAuthLink(t *testing.T) {
	m := newTestManager(t)
	_, err := m.CreateUser("oauth-user")
	require.NoError(t, err)

	token, err := m.LinkOAuth("oauth-user", auth.TokenKindGoogle, "google-subject-1", auth.OAuthSecret{
		Email: "user@example.com",
		Name:  "Example User",
	})
	require.NoError(t, err)
	require.NotNil(t, token.Secret.OAuth)
	assert.Nil(t, token.Secret.Password)
	assert.Equal(t, []byte("google-subject-1"), token.ExternalID)
	assert.False(t, token.Secret.OAuth.RefreshedAt.IsZero())

	// provider 가 준 ID 로 사용자를 역추적한다
	found, err := m.FindOAuthUser(auth.TokenKindGoogle, "google-subject-1")
	require.NoError(t, err)
	assert.Equal(t, "oauth-user", found.Name)

	// 같은 값이라도 provider 가 다르면 별개다
	_, err = m.FindOAuthUser(auth.TokenKindGitHub, "google-subject-1")
	assert.Error(t, err)

	_, err = m.LinkOAuth("oauth-user", auth.TokenKindGitHub, "google-subject-1", auth.OAuthSecret{})
	require.NoError(t, err)

	links, err := m.ListOAuth("oauth-user")
	require.NoError(t, err)
	assert.Len(t, links, 2)

	// 연결을 끊으면 역추적되지 않는다
	require.NoError(t, m.RevokeOAuth("oauth-user", auth.TokenKindGoogle, token.Index))

	_, err = m.FindOAuthUser(auth.TokenKindGoogle, "google-subject-1")
	assert.Error(t, err)

	links, err = m.ListOAuth("oauth-user")
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestOAuthLinkIsExclusive(t *testing.T) {
	m := newTestManager(t)
	for _, name := range []string{"first", "second"} {
		_, err := m.CreateUser(name)
		require.NoError(t, err)
	}

	_, err := m.LinkOAuth("first", auth.TokenKindGoogle, "shared-subject", auth.OAuthSecret{})
	require.NoError(t, err)

	// 같은 외부 계정이 두 사용자에 붙으면 안 된다
	_, err = m.LinkOAuth("second", auth.TokenKindGoogle, "shared-subject", auth.OAuthSecret{})
	assert.Error(t, err)
}

func TestOAuthProfileUpdate(t *testing.T) {
	m := newTestManager(t)
	_, err := m.CreateUser("profile-user")
	require.NoError(t, err)

	past := time.Now().Add(-24 * time.Hour)
	_, err = m.LinkOAuth("profile-user", auth.TokenKindGitHub, "gh-1", auth.OAuthSecret{
		Email:       "old@example.com",
		RefreshedAt: past,
	})
	require.NoError(t, err)

	require.NoError(t, m.UpdateOAuthProfile(auth.TokenKindGitHub, "gh-1", auth.OAuthSecret{
		Email: "new@example.com",
		Name:  "New Name",
	}))

	links, err := m.ListOAuth("profile-user")
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.NotNil(t, links[0].Secret.OAuth)
	assert.Equal(t, "new@example.com", links[0].Secret.OAuth.Email)
	assert.Equal(t, "New Name", links[0].Secret.OAuth.Name)
	assert.True(t, links[0].Secret.OAuth.RefreshedAt.After(past))

	// 없는 연결은 error 다
	assert.Error(t, m.UpdateOAuthProfile(auth.TokenKindGitHub, "no-such", auth.OAuthSecret{}))
}

func TestOAuthRejectsNonProvider(t *testing.T) {
	m := newTestManager(t)
	_, err := m.CreateUser("reject-user")
	require.NoError(t, err)

	// passkey 나 password 는 이 경로로 다루지 않는다
	_, err = m.LinkOAuth("reject-user", auth.TokenKindPasskey, "x", auth.OAuthSecret{})
	assert.Error(t, err)

	_, err = m.FindOAuthUser(auth.TokenKindPassword, "x")
	assert.Error(t, err)

	assert.Error(t, m.RevokeOAuth("reject-user", auth.TokenKindPassword, 0))

	// ID 가 비면 역추적할 수 없다
	_, err = m.LinkOAuth("reject-user", auth.TokenKindGoogle, "", auth.OAuthSecret{})
	assert.Error(t, err)
}
