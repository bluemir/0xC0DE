package gorm

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func (s *Store) CreateRoleBinding(roleBinding *auth.RoleBinding) error {
	return s.db.Create(roleBinding).Error
}
func (s *Store) GetRoleBinding(subject auth.Subject) (*auth.RoleBinding, error) {
	rb := &auth.RoleBinding{}
	if err := s.db.Where(auth.RoleBinding{
		Subject: auth.Subject{
			Kind: subject.Kind,
			Name: subject.Name,
		},
	}).Take(rb).Error; err != nil {
		return nil, err
	}

	return rb, nil
}
func (s *Store) ListRoleBinding() ([]auth.RoleBinding, error) {
	rbs := []auth.RoleBinding{}
	if err := s.db.Find(rbs).Error; err != nil {
		return nil, err
	}

	return rbs, nil
}
func (s *Store) UpdateRoleBinding(rb *auth.RoleBinding) error {
	return s.db.Save(rb).Error
}
func (s *Store) DeleteRoleBinding(subject auth.Subject) error {
	return s.db.Where(auth.RoleBinding{
		Subject: auth.Subject{
			Kind: subject.Kind,
			Name: subject.Name,
		},
	}).Delete(auth.RoleBinding{}).Error
}
