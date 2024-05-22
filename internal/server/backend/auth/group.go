package auth

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (m *Manager) CreateGroup(name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.store.CreateGroup(g); err != nil {
		return nil, err
	}
	return g, nil
}
func (m *Manager) EnsureGroup(name string) (*Group, error) {
	g, err := m.store.GetGroup(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			g, err := m.CreateGroup(name)
			if err != nil {
				return nil, err
			}
			return g, nil
		}
		return nil, err
	}
	return g, nil
}
func (m *Manager) ListGroup() ([]Group, error) {
	return m.store.ListGroup()
}
func (m *Manager) DeleteGroup(name string) error {
	return m.store.DeleteGroup(name)
}
