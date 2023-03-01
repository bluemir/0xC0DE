package static

import "github.com/bluemir/0xC0DE/internal/auth"

func (s *Store) CreateGroup(group *auth.Group) error {
	s.groups[group.Name] = *group
	return nil
}
func (s *Store) GetGroup(name string) (*auth.Group, error) {
	group := s.groups[name]
	return &group, nil
}
func (s *Store) ListGroup() ([]auth.Group, error) {
	groups := []auth.Group{}

	for _, group := range s.groups {
		groups = append(groups, group)
	}
	return groups, nil
}
func (s *Store) DeleteGroup(name string) error {
	delete(s.groups, name)
	return nil
}
