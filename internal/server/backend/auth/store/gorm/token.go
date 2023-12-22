package gorm

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func (s *Store) CreateToken(token *auth.Token) error {
	return s.db.Create(token).Error
}
func (s *Store) GetToken(username, hashedKey string) (*auth.Token, error) {
	token := &auth.Token{
		Username:  username,
		HashedKey: hashedKey,
	}
	if err := s.db.Take(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}
func (s *Store) ListToken(username string) ([]auth.Token, error) {
	tokens := []auth.Token{}
	if err := s.db.Find(tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}
func (s *Store) DeleteToken(username, revokeKey string) error {
	return s.db.Delete(&auth.Token{Username: username, RevokeKey: revokeKey}).Error
}
