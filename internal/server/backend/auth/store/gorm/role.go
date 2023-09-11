package gorm

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func (s *Store) CreateRole(role *auth.Role) error {
	return s.db.Create(role).Error
}
func (s *Store) GetRole(name string) (*auth.Role, error) {
	role := &auth.Role{}
	if err := s.db.Take(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}
func (s *Store) ListRole() ([]auth.Role, error) {
	roles := []auth.Role{}
	if err := s.db.Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}
func (s *Store) UpdateRole(role *auth.Role) error {
	return s.db.Save(role).Error
}
func (s *Store) DeleteRole(name string) error {
	return s.db.Delete(&auth.Role{Name: name}).Error
}
