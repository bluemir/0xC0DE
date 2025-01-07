package auth

import (
	"github.com/bluemir/0xC0DE/internal/functional"
	"github.com/pkg/errors"
)

type Assign struct {
	Subject Subject `gorm:"embedded;embeddedPrefix:subject_"`
	Role    string  `gorm:"primaryKey"`
}

type Subject struct {
	Kind string `gorm:"primaryKey;size:256" expr:"kind"`
	Name string `gorm:"primaryKey;size:256" expr:"name"`
}

func (m *Manager) AssignRole(subject Subject, roleName string) error {
	if err := m.db.FirstOrCreate(&Assign{
		Subject: subject,
		Role:    roleName,
	}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *Manager) DiscardRole(subject Subject, roleName string) error {
	if err := m.db.Where(Assign{
		Subject: subject,
		Role:    roleName,
	}).Delete(Assign{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (m *Manager) ListAssignedRole(subject Subject) ([]Role, error) {
	assigns := []Assign{}
	if err := m.db.Where("subject_kind = ?", subject.Kind).Where(`subject_name = "" OR subject_name = ?`, subject.Name).Find(&assigns).Error; err != nil {
		return nil, err
	}
	if subject.Kind == KindGroup {
		assigns = append(assigns, Assign{
			Subject: subject,
			Role:    subject.Name,
		})
	}

	roles, err := functional.MapWithError(assigns, func(assign Assign) (*Role, error) {
		role := &Role{}

		if err := m.db.Where(Role{Name: assign.Role}).Take(role).Error; err != nil {
			return nil, errors.WithStack(err)
		}
		return role, nil
	})
	if err != nil {
		return nil, err
	}

	return functional.Map(roles, func(r *Role) Role { return *r }), nil
}
