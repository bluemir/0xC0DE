package auth

func (m *Manager) CreateGroup(name string) (*Group, error) {
	g := &Group{
		Name: name,
	}
	if err := m.store.CreateGroup(g); err != nil {
		return nil, err
	}
	return g, nil
}
func (m *Manager) ListGroup() ([]Group, error) {
	return m.store.ListGroup()
}
func (m *Manager) DeleteGroup(name string) error {
	return m.store.DeleteGroup(name)
}
