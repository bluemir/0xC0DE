package auth

import (
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"gorm.io/gorm"
)

func (m *Manager) CreateRole(name string, rules []Rule) (*Role, error) {
	role := &Role{
		Name:  name,
		Rules: rules,
	}
	if err := m.store.CreateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}
func (m *Manager) GetRole(name string) (*Role, error) {
	return m.store.GetRole(name)
}
func (m *Manager) ListRole() ([]Role, error) {
	return m.store.ListRole()
}
func (m *Manager) UpdateRole(role *Role) error {
	return m.store.UpdateRole(role)
}
func (m *Manager) DeleteRole(name string) error {
	return m.store.DeleteRole(name)
}

func (m *Manager) AssignRole(subject Subject, roleName string) error {
	binding, err := m.store.GetRoleBinding(subject)
	if err != nil {
		return err
	}
	binding.RoleNames[roleName] = x

	if err := m.store.UpdateRoleBinding(binding); err != nil {
		return err
	}
	return nil
}

func (m *Manager) DiscardRole(subject Subject, roleName string) error {
	rb, err := m.store.GetRoleBinding(subject)
	if err != nil {
		rb = &RoleBinding{
			Subject:   subject,
			RoleNames: Set{roleName: x},
		}
		if err = m.store.CreateRoleBinding(rb); err != nil {
			return err
		}
	}

	rb.RoleNames[roleName] = x

	if err := m.store.UpdateRoleBinding(rb); err != nil {
		return err
	}
	return nil
}
func (m *Manager) ListAssignedRole(subject Subject) ([]Role, error) {
	roleNames := Set{}

	if rb, err := m.store.GetRoleBinding(subject); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	} else {
		maps.Copy(roleNames, rb.RoleNames)
	}

	if subject.Kind == "group" {
		roleNames.Add(subject.Name)
	}

	roles := []Role{}
	for roleName := range roleNames {
		role, err := m.store.GetRole(roleName)
		if err != nil {
			return nil, err
		}
		roles = append(roles, *role)
	}

	return roles, nil
}
