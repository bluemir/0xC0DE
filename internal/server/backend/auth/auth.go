package auth

import (
	"context"
	"net/http"
	"net/url"

	"gorm.io/gorm"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
)

type IManager interface {
	// Authn
	Default(ctx context.Context, username, unhashedKey string) (*User, error)
	HTTP(req *http.Request) (*User, error)

	// Authz
	Can(ctx context.Context, user *User, verb Verb, res Resource) bool

	// Shotcuts
	Register(ctx context.Context, username, unhashedKey string, opts ...CreateUserOption) (*User, *Token, error)

	// User
	CreateUser(ctx context.Context, username string, opts ...CreateUserOption) (*User, error)
	GetUser(ctx context.Context, username string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	ListUser(ctx context.Context, opts ...meta.ListOptionFn) ([]User, error)
	UpdateEmail(ctx context.Context, username string, email string) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, username string) error

	// Token
	IssueToken(ctx context.Context, username string, kind TokenKind, unhashedSecret string, opts ...TokenOpt) (*Token, error)
	GenerateAccessKey(ctx context.Context, username string, opts ...TokenOpt) (*Token, string, error)
	UpdatePassword(ctx context.Context, username string, password string) error
	GetToken(ctx context.Context, username string, kind TokenKind, index int) (*Token, error)
	ListToken(ctx context.Context, username string) ([]Token, error)
	RevokeToken(ctx context.Context, username string, kind TokenKind, index int) error

	// Recovery(계정 복구)
	IssueRecovery(ctx context.Context, username string) (string, error)
	ValidateRecovery(ctx context.Context, username string, unhashedSecret string) (*User, error)
	RevokeRecovery(ctx context.Context, username string) error

	// Verification(이메일 확인)
	IssueVerification(ctx context.Context, username string, email string) (string, error)
	ConfirmVerification(ctx context.Context, username string, unhashedSecret string) (*User, error)
	RevokeVerification(ctx context.Context, username string) error

	// Passkey(WebAuthn)
	BeginPasskeyRegistration(ctx context.Context, site *url.URL, username string) (*protocol.CredentialCreation, *PasskeyRegistration, error)
	FinishPasskeyRegistration(site *url.URL, reg *PasskeyRegistration, label string, req *http.Request) (*User, *Token, error)
	BeginPasskeyLogin(ctx context.Context, site *url.URL) (*protocol.CredentialAssertion, []byte, error)
	FinishPasskeyLogin(site *url.URL, sessionData []byte, req *http.Request) (*User, error)
	ListPasskey(ctx context.Context, username string) ([]Token, error)
	RevokePasskey(ctx context.Context, username string, index int) error

	// OAuth(social login)
	LinkOAuth(ctx context.Context, username string, kind TokenKind, externalID string, profile OAuthSecret) (*Token, error)
	FindOAuthUser(ctx context.Context, kind TokenKind, externalID string) (*User, error)
	UpdateOAuthProfile(ctx context.Context, kind TokenKind, externalID string, profile OAuthSecret) error
	ListOAuth(ctx context.Context, username string) ([]Token, error)
	RevokeOAuth(ctx context.Context, username string, kind TokenKind, index int) error

	// Group
	CreateGroup(ctx context.Context, name string) (*Group, error)
	ListGroup(ctx context.Context, opts ...meta.ListOptionFn) ([]Group, error)
	DeleteGroup(ctx context.Context, name string) error
	// group has only name. no need to update
	// adding or remove group member will be handle in user method

	// Role
	CreateRole(ctx context.Context, name string, rules []Rule) (*Role, error)
	GetRole(ctx context.Context, name string) (*Role, error)
	ListRole(ctx context.Context, opts ...meta.ListOptionFn) ([]Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, name string) error

	// RoleBinding
	AssignRole(ctx context.Context, subject Subject, roleName string) error
	DiscardRole(ctx context.Context, subject Subject, roleName string) error
	ListAssignedRole(ctx context.Context, subject Subject) ([]Role, error)
}

var _ IManager = (*Manager)(nil)

// Config from file
type Config struct {
	Salt    string
	Passkey PasskeyConfig
}

type Manager struct {
	db      *gorm.DB
	salt    string
	passkey PasskeyConfig
	events  *pubsub.Hub
}

func New(ctx context.Context, conf *Config, db *gorm.DB, events *pubsub.Hub) (*Manager, error) {

	if err := db.AutoMigrate(
		&User{},
		&Group{},
		&ServiceAccount{},
		&Token{},
		&Recovery{},
		&Verification{},
		&Role{},
		&Assign{},
	); err != nil {
		return nil, errors.WithStack(err)
	}

	m := &Manager{db, conf.Salt, conf.Passkey, events}

	if err := m.initialize(ctx); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) Default(ctx context.Context, username, unhashedSecret string) (*User, error) {
	token, err := m.GetToken(ctx, username, TokenKindPassword, 0)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if err := token.Validate(unhashedSecret); err != nil {
		return nil, err
	}
	return m.GetUser(ctx, username)
}

func (m *Manager) Can(ctx context.Context, user *User, verb Verb, res Resource) bool {
	logger := logrus.WithField("verb", verb).WithField("resource", res)
	if user != nil {
		logger = logger.WithField("user", user.Name).WithField("group", user.Groups)
	}

	//logger.Tracef("length of subject: %d", len(user.Subjects()))

	for _, subject := range user.Subjects() {
		logger := logger.WithField("subject", subject)

		roles, err := m.ListAssignedRole(ctx, subject)
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
