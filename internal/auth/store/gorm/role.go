package gorm

import (
	"encoding/json"

	"github.com/bluemir/0xC0DE/internal/auth"
)

type Role struct {
	Name string
	Data string
}

func EncodeRole(role *auth.Role) (*Role, error) {
	buf, err := json.Marshal(role)
	if err != nil {
		return nil, err
	}
	return &Role{
		Name: role.Name,
		Data: string(buf),
	}, nil
}
func DecodeRole(rw *Role) (*auth.Role, error) {
	role := auth.Role{}
	if err := json.Unmarshal([]byte(rw.Data), &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *Store) CreateRole(role *auth.Role) error {
	r, err := EncodeRole(role)
	if err != nil {
		return err
	}
	return s.db.Create(r).Error
}
func (s *Store) GetRole(name string) (*auth.Role, error) {
	rw := &Role{}
	if err := s.db.Take(rw).Error; err != nil {
		return nil, err
	}
	role, err := DecodeRole(rw)
	if err != nil {
		return nil, err
	}

	return role, nil
}
func (s *Store) ListRole() ([]auth.Role, error) {
	rs := []Role{}
	if err := s.db.Find(&rs).Error; err != nil {
		return nil, err
	}
	roles := []auth.Role{}
	for _, r := range rs {
		role, err := DecodeRole(&r)
		if err != nil {
			return nil, err
		}
		roles = append(roles, *role)
	}
	return roles, nil
}
func (s *Store) UpdateRole(role *auth.Role) error {
	r, err := EncodeRole(role)
	if err != nil {
		return err
	}
	return s.db.Save(r).Error
}
func (s *Store) DeleteRole(name string) error {
	return s.db.Delete(&Role{Name: name}).Error
}
