package gorm

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

// User CRUD
func (s *Store) CreateUser(user *auth.User) error {
	return s.db.Create(user).Error
}

func (s *Store) GetUser(username string) (*auth.User, error) {
	user := &auth.User{Name: username}
	if err := s.db.Take(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
func (s *Store) ListUser() ([]auth.User, error) {
	users := []auth.User{}
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
func (s *Store) UpdateUser(user *auth.User) error {
	return s.db.Save(user).Error
}

func (s *Store) DeleteUser(username string) error {
	return s.db.Delete(&auth.User{Name: username}).Error
}
