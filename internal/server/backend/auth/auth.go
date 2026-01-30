package auth

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
)

type IManager interface {
	// Authn
	Default(username, unhashedKey string) (*User, error)
	HTTP(req *http.Request) (*User, error)

	// Authz
	Can(user *User, verb Verb, res Resource) bool

	// Shotcuts
	Register(username, unhashedKey string, opts ...CreateUserOption) (*User, *Token, error)

	// User
	CreateUser(username string, opts ...CreateUserOption) (*User, error)
	GetUser(username string) (*User, error)
	ListUser(opts ...meta.ListOptionFn) ([]User, error)
	UpdateUser(user *User) error
	DeleteUser(username string) error

	// Token
	IssueToken(username string, kind TokenKind, unhashedSecret string, opts ...TokenOpt) (*Token, error)
	GenerateAccessKey(username string, opts ...TokenOpt) (*Token, string, error)
	UpdatePassword(username string, password string) error
	GetToken(username string, kind TokenKind, index int) (*Token, error)
	ListToken(username string) ([]Token, error)
	RevokeToken(username string, kind TokenKind, index int) error

	// Group
	CreateGroup(name string) (*Group, error)
	ListGroup(opts ...meta.ListOptionFn) ([]Group, error)
	DeleteGroup(name string) error
	// group has only name. no need to update
	// adding or remove group member will be handle in user method

	// Role
	CreateRole(name string, rules []Rule) (*Role, error)
	GetRole(name string) (*Role, error)
	ListRole(opts ...meta.ListOptionFn) ([]Role, error)
	UpdateRole(role *Role) error
	DeleteRole(name string) error

	// RoleBinding
	AssignRole(subject Subject, roleName string) error
	DiscardRole(subject Subject, roleName string) error
	ListAssignedRole(Subject) ([]Role, error)
}

var _ IManager = (*Manager)(nil)

type Manager struct {
	db   *gorm.DB
	salt string
}

func New(db *gorm.DB, salt string) (*Manager, error) {

	if err := db.AutoMigrate(
		&User{},
		&Group{},
		&ServiceAccount{},
		&Token{},
		&Role{},
		&Assign{},
	); err != nil {
		return nil, errors.WithStack(err)
	}

	m := &Manager{db, salt}

	if err := m.initialize(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) Default(username, unhashedSecret string) (*User, error) {
	token, err := m.GetToken(username, TokenKindPassword, 0)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if err := token.Validate(unhashedSecret); err != nil {
		return nil, err
	}
	return m.GetUser(username)
}

func (m *Manager) Can(user *User, verb Verb, res Resource) bool {
	logger := logrus.WithField("verb", verb).WithField("resource", res)
	if user != nil {
		logger = logger.WithField("user", user.Name).WithField("group", user.Groups)
	}

	//logger.Tracef("length of subject: %d", len(user.Subjects()))

	for _, subject := range user.Subjects() {
		logger := logger.WithField("subject", subject)

		roles, err := m.ListAssignedRole(subject)
		if err != nil {
			logger.Warn(err)
			continue // skip
		}

		for _, role := range roles {
			logger := logger.WithField("role", role.Name)

			if role.IsAllow(Context{
				User:     user,
				Subject:  subject,
				Verb:     verb,
				Resource: res,
			}) {
				logger.Tracef("allowed")
				return true
			}
			logger.Trace("next")
		}
	}

	logger.Trace("rejected")
	return false
}
