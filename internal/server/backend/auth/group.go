package auth

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type Group struct {
	Name string `gorm:"primaryKey;size:256"`
}

func (m *Manager) CreateGroup(ctx context.Context, name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.db.WithContext(ctx).Create(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return g, nil
}
func (m *Manager) EnsureGroup(ctx context.Context, name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.db.WithContext(ctx).FirstOrCreate(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := m.db.WithContext(ctx).Save(g).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return g, nil
}
func (m *Manager) ListGroup(ctx context.Context, opts ...meta.ListOptionFn) ([]Group, error) {
	opt := meta.ListOption{
		Limit:  -1,
		Offset: 0,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListGroupWithOption(ctx, opt)
}
func (m *Manager) ListGroupWithOption(ctx context.Context, option meta.ListOption) ([]Group, error) {
	groups := []Group{}

	if err := m.db.WithContext(ctx).Offset(option.Offset).Limit(option.Limit).Find(&groups).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return groups, nil
}
func (m *Manager) DeleteGroup(ctx context.Context, name string) error {
	if err := m.db.WithContext(ctx).Delete(Group{Name: name}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
