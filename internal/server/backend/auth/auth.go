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
	Default(username, unhashedKey string) (*User, error)
	HTTP(req *http.Request) (*User, error)

	// Authz
	Can(user *User, verb Verb, res Resource) bool

	// Shotcuts
	Register(username, unhashedKey string, opts ...CreateUserOption) (*User, *Token, error)

	// User
	CreateUser(username string, opts ...CreateUserOption) (*User, error)
	GetUser(username string) (*User, error)
	FindUserByEmail(email string) (*User, error)
	ListUser(opts ...meta.ListOptionFn) ([]User, error)
	UpdateEmail(username string, email string) error
	UpdateUser(user *User) error
	DeleteUser(username string) error

	// Token
	IssueToken(username string, kind TokenKind, unhashedSecret string, opts ...TokenOpt) (*Token, error)
	GenerateAccessKey(username string, opts ...TokenOpt) (*Token, string, error)
	UpdatePassword(username string, password string) error
	GetToken(username string, kind TokenKind, index int) (*Token, error)
	ListToken(username string) ([]Token, error)
	RevokeToken(username string, kind TokenKind, index int) error

	// Recovery(계정 복구)
	IssueRecovery(username string) (string, error)
	ValidateRecovery(username string, unhashedSecret string) (*User, error)
	RevokeRecovery(username string) error

	// Verification(이메일 확인)
	IssueVerification(username string, email string) (string, error)
	ConfirmVerification(username string, unhashedSecret string) (*User, error)
	RevokeVerification(username string) error

	// Passkey(WebAuthn)
	BeginPasskeyRegistration(site *url.URL, username string) (*protocol.CredentialCreation, *PasskeyRegistration, error)
	FinishPasskeyRegistration(site *url.URL, reg *PasskeyRegistration, label string, req *http.Request) (*User, *Token, error)
	BeginPasskeyLogin(site *url.URL) (*protocol.CredentialAssertion, []byte, error)
	FinishPasskeyLogin(site *url.URL, sessionData []byte, req *http.Request) (*User, error)
	ListPasskey(username string) ([]Token, error)
	RevokePasskey(username string, index int) error

	// OAuth(social login)
	LinkOAuth(username string, kind TokenKind, externalID string, profile OAuthSecret) (*Token, error)
	FindOAuthUser(kind TokenKind, externalID string) (*User, error)
	UpdateOAuthProfile(kind TokenKind, externalID string, profile OAuthSecret) error
	ListOAuth(username string) ([]Token, error)
	RevokeOAuth(username string, kind TokenKind, index int) error

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

	// processContext 는 요청에 매이지 않은 일에 쓸 context 다.
	// 계정 삭제 이벤트처럼 호출 경로가 context 를 들고 오지 않는 자리에서 쓴다.
	processContext context.Context
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

	m := &Manager{db, conf.Salt, conf.Passkey, events, ctx}

	// 계정이 지워지면 딸린 행도 지워야 한다. 자기 이벤트를 자기가 받는 모양이지만,
	// 같은 자리에 다른 백엔드의 핸들러도 걸리므로 치우는 쪽을 한 줄로 모아 둔다.
	events.AddHandler(EventUserDeleted{}, &userCleanup{manager: m})

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
