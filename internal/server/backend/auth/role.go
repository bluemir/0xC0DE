package auth

import (
	"github.com/bluemir/functional"
	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type Role struct {
	Name  string `gorm:"primaryKey;size:256"`
	Rules []Rule `gorm:"type:bytes;serializer:gob"`
}

type Rule struct {
	Verbs      []Verb      `gorm:"type:bytes;serializer:gob"`
	Selector   KeyValues   `gorm:"type:bytes;serializer:gob"`
	Conditions []Condition `gorm:"type:bytes;serializer:gob"`
	//
	// Attribute check?
	// eg)
	// user
	//   subject.kind == "user" && subject.name == object.name
	// project
	//   subject == object.owner
}

func (m *Manager) CreateRole(name string, rules []Rule) (*Role, error) {
	role := &Role{
		Name:  name,
		Rules: rules,
	}
	if err := m.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}
func (m *Manager) GetRole(name string) (*Role, error) {
	role := Role{}
	if err := m.db.Where(Role{Name: name}).Take(&role).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &role, nil
}
func (m *Manager) ListRole(opts ...meta.ListOptionFn) ([]Role, error) {
	opt := meta.ListOption{}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListRoleWithOption(&opt)
}
func (m *Manager) ListRoleWithOption(option *meta.ListOption) ([]Role, error) {
	if option.Limit == 0 {
		option.Limit = 20
	}
	roles := []Role{}

	if err := m.db.Offset(option.Offset).Limit(option.Limit).Find(&roles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return roles, nil
}
func (m *Manager) UpdateRole(role *Role) error {
	if err := m.db.Save(role).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (m *Manager) DeleteRole(name string) error {
	if err := m.db.Where(Role{Name: name}).Delete(Role{Name: name}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (role *Role) IsAllow(ctx Context) bool {
	logrus.Trace(ctx)

	return functional.Some(role.Rules, func(rule Rule) bool {
		return rule.IsMatched(ctx)
	})
}
func (rule *Rule) IsMatched(ctx Context) bool {
	if !(functional.Contain(rule.Verbs, ctx.Verb) || len(rule.Verbs) == 0) {
		return false
	}

	if !rule.Selector.IsSubsetOf(ctx.Resource) {
		return false
	}

	for _, cond := range rule.Conditions {
		r, err := cond.IsMatched(ctx)
		if err != nil {
			logrus.Error(err)
			return false
		}
		if !r {
			return false
		}
	}

	return true
}
