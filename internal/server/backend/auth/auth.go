package auth

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type IManager interface {
	// Authn
	Default(username, unhashedKey string) (*User, error)
	HTTP(req *http.Request) (*User, error)

	// Authz
	Can(user *User, verb Verb, res Resource) bool

	// Shotcuts
	Register(username, unhashedKey string) (*User, *Token, error)

	// User
	CreateUser(username string, opts ...CreateUserOption) (*User, error)
	GetUser(username string) (*User, error)
	ListUser() ([]User, error)
	UpdateUser(user *User) error
	DeleteUser(username string) error

	// Token
	IssueToken(username, unhashedKey string, opts ...TokenOpt) (*Token, error)
	GenerateToken(username string, opts ...TokenOpt) (*Token, string, error)
	GetToken(username, unhashedKey string) (*Token, error)
	ListToken(username string) ([]Token, error)
	RevokeToken(username, unhashedKey string) error

	// Group
	CreateGroup(name string) (*Group, error)
	ListGroup() ([]Group, error)
	DeleteGroup(name string) error
	// group has only name. no need to update
	// adding or remove group member will be handle in user method

	// Role
	CreateRole(name string, rules []Rule) (*Role, error)
	GetRole(name string) (*Role, error)
	ListRole() ([]Role, error)
	UpdateRole(role *Role) error
	DeleteRole(name string) error

	// RoleBinding
	AssignRole(subject Subject, roleName string) error
	DiscardRole(subject Subject, roleName string) error
	ListAssignedRole(Subject) ([]Role, error)
}
type Manager struct {
	store AuthStore
	salt  string
}

func New(store AuthStore, salt string) (*Manager, error) {
	return &Manager{store, salt}, nil
}

func (m *Manager) Default(username, unhashedKey string) (*User, error) {
	if _, err := m.GetToken(username, unhashedKey); err != nil {
		return nil, ErrUnauthorized
	}
	return m.store.GetUser(username)
}

func (m *Manager) Can(user *User, verb Verb, res Resource) bool {
	for _, subject := range user.Subjects() {
		roles, err := m.ListAssignedRole(subject)
		if err != nil {
			logrus.Warn(err)
			continue // skip
		}
		logrus.Tracef("%#v", roles)

		for _, role := range roles {
			if role.IsAllow(Context{
				User:     user,
				Subject:  subject,
				Verb:     verb,
				Resource: res,
			}) {
				return true
			}
		}
	}

	return false
}

/*
func (m *Manager) Could(user *User, verb Verb, res Resource) (bool, error) {
	// could collect errors

	for _, subject := range user.Subjects() {
		roles, err := m.ListAssignedRole(subject)
		if err != nil && exitOnErr {
			return false, err
		}
		for _, role := range roles {
			if role.IsAllow(verb, res) {
				return true, nil
			}
		}

	}

	return false, nil
}
*/
