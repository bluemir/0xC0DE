package auth

import (
	"github.com/cockroachdb/errors"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type Group struct {
	Name string `gorm:"primaryKey;size:256"`
}

func (m *Manager) CreateGroup(name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.db.Create(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return g, nil
}
func (m *Manager) EnsureGroup(name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.db.FirstOrCreate(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := m.db.Save(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return g, nil
}
func (m *Manager) ListGroup(opts ...meta.ListOptionFn) ([]Group, error) {
	opt := meta.ListOption{
		Limit:  -1,
		Offset: 0,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListGroupWithOption(opt)
}
func (m *Manager) ListGroupWithOption(option meta.ListOption) ([]Group, error) {
	groups := []Group{}

	if err := m.db.Offset(option.Offset).Limit(option.Limit).Find(&groups).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return groups, nil
}
func (m *Manager) DeleteGroup(name string) error {
	if err := m.db.Delete(Group{Name: name}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
