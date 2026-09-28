package auth

import (
	"github.com/cockroachdb/errors"
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

	if len(assigns) == 0 {
		return []Role{}, nil
	}

	roleNames := make([]string, 0, len(assigns))
	for _, assign := range assigns {
		roleNames = append(roleNames, assign.Role)
	}

	roles := []Role{}
	if err := m.db.Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return roles, nil
}
