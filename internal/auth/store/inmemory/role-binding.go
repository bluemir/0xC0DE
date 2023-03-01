package static

import "github.com/bluemir/0xC0DE/internal/auth"

func (s *Store) CreateRoleBinding(rolebinding *auth.RoleBinding) error {
	s.rolebindings[rolebinding.Subject] = *rolebinding
	return nil
}
func (s *Store) GetRoleBinding(subject auth.Subject) (*auth.RoleBinding, error) {
	rb := s.rolebindings[subject]
	return &rb, nil
}
func (s *Store) ListRoleBinding() ([]auth.RoleBinding, error) {
	rbs := []auth.RoleBinding{}

	for _, rb := range rbs {
		rbs = append(rbs, rb)
	}
	return rbs, nil
}
func (s *Store) UpdateRoleBinding(rb *auth.RoleBinding) error {
	s.rolebindings[rb.Subject] = *rb
	return nil
}
func (s *Store) DeleteRoleBinding(subject auth.Subject) error {
	delete(s.rolebindings, subject)
	return nil
}
