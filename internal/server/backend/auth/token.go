package auth

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
)

type Token struct {
	Username  string `gorm:"primaryKey;size:256"`
	HashedKey string `gorm:"primaryKey;size:256"`
	RevokeKey string
	ExpiredAt *time.Time
}

func (m *Manager) IssueToken(username, unhashedKey string, opts ...TokenOpt) (*Token, error) {
	user, err := m.GetUser(username)
	if err != nil {
		return nil, err
	}

	token := &Token{
		Username:  username,
		HashedKey: Hash(unhashedKey, user.Salt, m.salt),
		RevokeKey: xid.New().String(),
	}
	for _, fn := range opts {
		fn(token)
	}

	if err := m.db.Create(token).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return token, nil
}
func (m *Manager) GenerateToken(username string, opts ...TokenOpt) (*Token, string, error) {
	unhashedKey := Hash(xid.New().String(), m.salt)

	t, err := m.IssueToken(username, unhashedKey, opts...)
	if err != nil {
		return nil, "", err
	}
	return t, unhashedKey, nil
}
func (m *Manager) GetToken(username, unhashedKey string) (*Token, error) {
	user, err := m.GetUser(username)
	if err != nil {
		return nil, err
	}
	hashedKey := Hash(unhashedKey, user.Salt, m.salt)

	token := Token{}
	if err := m.db.Where(Token{Username: username, HashedKey: hashedKey}).Take(&token).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &token, nil
}
func (m *Manager) ListToken(username string) ([]Token, error) {
	tokens := []Token{}

	if err := m.db.Where(Token{Username: username}).Find(&tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return tokens, nil
}
func (m *Manager) RevokeToken(username, revokeKey string) error {
	if err := m.db.Where(Token{Username: username, RevokeKey: revokeKey}).Delete(Token{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
