package auth

import (
	"github.com/rs/xid"
)

func (m *Manager) Register(username, unhashedKey string) (*User, *Token, error) {
	user, err := m.CreateUser(username)
	if err != nil {
		return nil, nil, err
	}

	token, err := m.IssueToken(username, unhashedKey)
	if err != nil {
		return user, nil, err
	}

	return user, token, nil
}

func (m *Manager) CreateUser(username string) (*User, error) {
	u := User{
		Name:   username,
		Salt:   xid.New().String(),
		Groups: Set{},
	}
	if err := m.store.CreateUser(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
func (m *Manager) GetUser(username string) (*User, error) {
	return m.store.GetUser(username)
}
func (m *Manager) ListUser() ([]User, error) {
	return m.store.ListUser()
}
func (m *Manager) UpdateUser(user *User) error {
	return m.store.UpdateUser(user)
}
func (m *Manager) DeleteUser(username string) error {
	return m.store.DeleteUser(username)
}
