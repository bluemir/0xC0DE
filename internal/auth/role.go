package auth

import "github.com/sirupsen/logrus"

func (m *Manager) IsAllow(resource Resource, verb Verb, user *User) bool {
	roles, err := m.GetBindingRoles(user)
	if err != nil {
		logrus.Warn(err)
	}
	for _, role := range roles {
		if role.IsAllow(resource, verb) {
			return true
		}
	}

	return false
}
func (r *Role) IsAllow(resource Resource, verb Verb) bool {
	for _, rule := range r.Rules {
		for _, v := range rule.Verbs {
			if rule.Resource.isSubsetOf(resource) && v == verb {
				return true
			}
		}
	}
	return false
}
func (m *Manager) GetBindingRoles(user *User) ([]Role, error) {
	//roles := []Role{}
	//if err := m.db.Preload("Roles").Where(user).Find(user).Error; err != nil {
	//	return nil, err
	//}
	//return user.Roles, nil
	// TODO handle Group

	bindings := []RoleBinding{}
	if err := m.db.Where(RoleBinding{
		User: user.Name,
	}).Find(bindings).Error; err != nil {
		return nil, err
	}

	result := []Role{}

	for _, b := range bindings {
		result = append(result, m.roles[b.Role])
	}

	return result, nil
}
func (m *Manager) BindRole(user *User, role string) error {
	if err := m.db.Save(&RoleBinding{
		User: user.Name,
		Role: role,
	}).Error; err != nil {
		return err
	}
	return nil
}
func (m *Manager) SetRole(role Role) {
	m.roles[role.Name] = role
}
func (m *Manager) DeleteRole(roleName string) {
	delete(m.roles, roleName)
}
