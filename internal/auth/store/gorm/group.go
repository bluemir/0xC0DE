package gorm

import "github.com/bluemir/0xC0DE/internal/auth"

func (s *Store) CreateGroup(group *auth.Group) error {
	return s.db.Create(group).Error
}
func (s *Store) GetGroup(name string) (*auth.Group, error) {
	g := &auth.Group{Name: name}
	if err := s.db.Take(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}
func (s *Store) ListGroup() ([]auth.Group, error) {
	gs := []auth.Group{}
	if err := s.db.Find(&gs).Error; err != nil {
		return nil, err
	}
	return gs, nil
}
func (s *Store) DeleteGroup(name string) error {
	return s.db.Delete(auth.Group{Name: name}).Error
}
