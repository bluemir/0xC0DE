package auth

import (
	"time"

	"github.com/cockroachdb/errors"
)

// OAuth provider 로 로그인하는 흐름은 아직 없다.
// provider 와 주고받는 부분(authorization code 교환, 프로필 조회)이 붙기 전까지,
// 여기서는 "외부 계정과 사용자의 연결" 만 다룬다.
//
// 연결은 Token 한 행이다. Kind 가 provider 를, ExternalID 가 provider 가 준 사용자 ID 를 담는다.
// 둘을 묶어 unique index 가 걸려 있어 같은 외부 계정이 두 사용자에 붙지 못한다.

// IsOAuth 는 Kind 가 OAuth provider 를 가리키는지 본다.
func (kind TokenKind) IsOAuth() bool {
	switch kind {
	case TokenKindGoogle, TokenKindGitHub:
		return true
	default:
		return false
	}
}

// LinkOAuth 는 외부 계정을 사용자에 연결한다.
// 이미 다른 사용자에 붙어 있는 계정이면 unique index 에 걸려 실패한다.
func (m *Manager) LinkOAuth(username string, kind TokenKind, externalID string, profile OAuthSecret) (*Token, error) {
	if !kind.IsOAuth() {
		return nil, errors.Errorf("token kind %q is not an oauth provider", kind)
	}
	if externalID == "" {
		return nil, errors.Errorf("oauth external id is empty")
	}

	if profile.RefreshedAt.IsZero() {
		profile.RefreshedAt = time.Now()
	}

	token := &Token{
		Username:   username,
		Kind:       kind,
		ExternalID: []byte(externalID),
		Secret:     Secret{OAuth: &profile},
	}
	if err := m.createToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// FindOAuthUser 는 provider 가 준 사용자 ID 로 연결된 사용자를 찾는다.
// 로그인 흐름이 붙으면 여기서 얻은 사용자를 세션에 담게 된다.
func (m *Manager) FindOAuthUser(kind TokenKind, externalID string) (*User, error) {
	if !kind.IsOAuth() {
		return nil, errors.Errorf("token kind %q is not an oauth provider", kind)
	}

	token := Token{}
	if err := m.db.Where("kind = ? AND external_id = ?", kind, []byte(externalID)).
		Take(&token).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return m.GetUser(token.Username)
}

// UpdateOAuthProfile 은 연결에 붙은 프로필을 다시 저장한다.
// provider 에서 프로필을 새로 받아왔을 때 쓴다.
func (m *Manager) UpdateOAuthProfile(kind TokenKind, externalID string, profile OAuthSecret) error {
	if !kind.IsOAuth() {
		return errors.Errorf("token kind %q is not an oauth provider", kind)
	}

	if profile.RefreshedAt.IsZero() {
		profile.RefreshedAt = time.Now()
	}

	// serializer 가 붙은 필드라 Update 가 아니라 구조체 Updates 를 쓴다.
	result := m.db.Model(&Token{}).
		Where("kind = ? AND external_id = ?", kind, []byte(externalID)).
		Updates(Token{Secret: Secret{OAuth: &profile}})
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.Errorf("oauth link not found. provider: %s", kind)
	}

	return nil
}

// ListOAuth 는 사용자에 연결된 외부 계정을 모두 돌려준다.
func (m *Manager) ListOAuth(username string) ([]Token, error) {
	tokens := []Token{}
	if err := m.db.Where("username = ? AND kind IN ?", username, []TokenKind{TokenKindGoogle, TokenKindGitHub}).
		Order("kind asc, `index` asc").
		Find(&tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return tokens, nil
}

// RevokeOAuth 는 연결을 끊는다.
func (m *Manager) RevokeOAuth(username string, kind TokenKind, index int) error {
	if !kind.IsOAuth() {
		return errors.Errorf("token kind %q is not an oauth provider", kind)
	}
	return m.RevokeToken(username, kind, index)
}
