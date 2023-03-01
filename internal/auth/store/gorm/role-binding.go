package gorm

import "github.com/bluemir/0xC0DE/internal/auth"

type RoleBinding struct {
	Kind      string
	Name      string
	RoleNames auth.Set
}

func EncodeRoleBinding(rb *auth.RoleBinding) (*RoleBinding, error) {
	return &RoleBinding{
		Kind:      rb.Subject.Kind,
		Name:      rb.Subject.Name,
		RoleNames: rb.RoleNames,
	}, nil
}
func DecodeRoleBinding(rb *RoleBinding) (*auth.RoleBinding, error) {
	return &auth.RoleBinding{
		Subject: auth.Subject{
			Kind: rb.Kind,
			Name: rb.Name,
		},
		RoleNames: rb.RoleNames,
	}, nil
}

func (s *Store) CreateRoleBinding(roleBinding *auth.RoleBinding) error {
	rb, err := EncodeRoleBinding(roleBinding)
	if err != nil {
		return err
	}
	return s.db.Create(rb).Error
}
func (s *Store) GetRoleBinding(subject auth.Subject) (*auth.RoleBinding, error) {
	rb := &RoleBinding{}
	if err := s.db.Where(RoleBinding{
		Kind: subject.Kind,
		Name: subject.Name,
	}).Take(rb).Error; err != nil {
		return nil, err
	}
	rolebinding, err := DecodeRoleBinding(rb)
	if err != nil {
		return nil, err
	}

	return rolebinding, nil
}
func (s *Store) ListRoleBinding() ([]auth.RoleBinding, error) {
	rbs := []RoleBinding{}
	if err := s.db.Find(rbs).Error; err != nil {
		return nil, err
	}
	rolebindings := []auth.RoleBinding{}
	for _, rb := range rbs {
		roleBinding, err := DecodeRoleBinding(&rb)
		if err != nil {
			return nil, err
		}
		rolebindings = append(rolebindings, *roleBinding)
	}
	return rolebindings, nil
}
func (s *Store) UpdateRoleBinding(rb *auth.RoleBinding) error {
	rolebinding, err := EncodeRoleBinding(rb)
	if err != nil {
		return err
	}
	return s.db.Save(rolebinding).Error
}
func (s *Store) DeleteRoleBinding(subject auth.Subject) error {
	return s.db.Where(RoleBinding{
		Kind: subject.Kind,
		Name: subject.Name,
	}).Delete(RoleBinding{}).Error
}
