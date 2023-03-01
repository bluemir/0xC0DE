package static

import "github.com/bluemir/0xC0DE/internal/auth"

func (s *Store) CreateRole(role *auth.Role) error {
	s.roles[role.Name] = *role
	return nil
}
func (s *Store) GetRole(name string) (*auth.Role, error) {
	role := s.roles[name]
	return &role, nil
}
func (s *Store) ListRole() ([]auth.Role, error) {
	roles := []auth.Role{}
	for _, role := range s.roles {
		roles = append(roles, role)
	}
	return roles, nil
}
func (s *Store) UpdateRole(role *auth.Role) error {
	s.roles[role.Name] = *role
	return nil
}
func (s *Store) DeleteRole(name string) error {
	delete(s.roles, name)
	return nil
}
