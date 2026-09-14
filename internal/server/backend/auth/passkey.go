package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PasskeyConfig 는 WebAuthn Relying Party 설정이다.
// ID, Origins 를 비워두면 요청이 들어온 사이트 주소에서 채운다.
type PasskeyConfig struct {
	DisplayName string
	ID          string
	Origins     []string
}

// handleSize 는 WebAuthn user handle 의 길이다. 규격 최대치는 64 byte 다.
const handleSize = 32

func newPasskeyHandle() ([]byte, error) {
	handle := make([]byte, handleSize)
	if _, err := rand.Read(handle); err != nil {
		return nil, errors.WithStack(err)
	}
	return handle, nil
}

// PasskeyRegistration 은 등록 ceremony 의 중간 상태다.
// begin 과 finish 사이에 사용자 세션에 담아 둔다.
type PasskeyRegistration struct {
	Username string
	Handle   []byte
	Session  []byte // webauthn.SessionData JSON
}

// passkeyIdentity 는 webauthn.User 구현체다.
// 아직 저장되지 않은(passkey 로 가입 중인) 사용자도 표현한다.
type passkeyIdentity struct {
	username    string
	handle      []byte
	credentials []webauthn.Credential
}

func (id *passkeyIdentity) WebAuthnID() []byte                         { return id.handle }
func (id *passkeyIdentity) WebAuthnName() string                       { return id.username }
func (id *passkeyIdentity) WebAuthnDisplayName() string                { return id.username }
func (id *passkeyIdentity) WebAuthnCredentials() []webauthn.Credential { return id.credentials }

func (m *Manager) relyingParty(site *url.URL) (*webauthn.WebAuthn, error) {
	conf := webauthn.Config{
		RPDisplayName: m.passkey.DisplayName,
		RPID:          m.passkey.ID,
		RPOrigins:     m.passkey.Origins,
	}

	if conf.RPID == "" {
		conf.RPID = site.Hostname()
	}
	if len(conf.RPOrigins) == 0 {
		conf.RPOrigins = []string{site.Scheme + "://" + site.Host}
	}
	if conf.RPDisplayName == "" {
		conf.RPDisplayName = conf.RPID
	}

	rp, err := webauthn.New(&conf)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rp, nil
}

// identity 는 username 으로 webauthn.User 를 만든다.
// 없는 사용자면 새 handle 을 만들어 가입에 쓸 수 있게 한다.
func (m *Manager) identity(username string) (*passkeyIdentity, error) {
	user, err := m.GetUser(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		handle, err := newPasskeyHandle()
		if err != nil {
			return nil, err
		}
		return &passkeyIdentity{username: username, handle: handle}, nil
	}
	if err != nil {
		return nil, err
	}

	// ID/PW 로 먼저 가입한 사용자는 handle 이 없을 수 있다.
	if len(user.Handle) == 0 {
		handle, err := newPasskeyHandle()
		if err != nil {
			return nil, err
		}
		if err := m.db.Model(user).Update("handle", handle).Error; err != nil {
			return nil, errors.WithStack(err)
		}
		user.Handle = handle
	}

	tokens, err := m.ListPasskey(username)
	if err != nil {
		return nil, err
	}

	credentials := []webauthn.Credential{}
	for _, token := range tokens {
		credential, err := token.PasskeyCredential()
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, *credential)
	}

	return &passkeyIdentity{
		username:    username,
		handle:      user.Handle,
		credentials: credentials,
	}, nil
}

// BeginPasskeyRegistration 은 passkey 등록 ceremony 를 시작한다.
// 반환된 PasskeyRegistration 은 FinishPasskeyRegistration 에 그대로 넘겨야 한다.
func (m *Manager) BeginPasskeyRegistration(site *url.URL, username string) (*protocol.CredentialCreation, *PasskeyRegistration, error) {
	rp, err := m.relyingParty(site)
	if err != nil {
		return nil, nil, err
	}

	identity, err := m.identity(username)
	if err != nil {
		return nil, nil, err
	}

	// 같은 authenticator 에 중복 등록되지 않게 이미 가진 credential 을 제외한다.
	exclusions := []protocol.CredentialDescriptor{}
	for _, credential := range identity.credentials {
		exclusions = append(exclusions, credential.Descriptor())
	}

	creation, session, err := rp.BeginRegistration(identity,
		// username 없이 로그인하려면 authenticator 가 credential 을 직접 들고 있어야 한다.
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(exclusions),
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	buf, err := json.Marshal(session)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return creation, &PasskeyRegistration{
		Username: identity.username,
		Handle:   identity.handle,
		Session:  buf,
	}, nil
}

// FinishPasskeyRegistration 은 authenticator 응답을 검증하고 passkey 를 저장한다.
// 아직 없는 사용자면 이 시점에 계정을 만든다.
func (m *Manager) FinishPasskeyRegistration(site *url.URL, reg *PasskeyRegistration, label string, req *http.Request) (*User, *Token, error) {
	rp, err := m.relyingParty(site)
	if err != nil {
		return nil, nil, err
	}

	session := webauthn.SessionData{}
	if err := json.Unmarshal(reg.Session, &session); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	identity := &passkeyIdentity{username: reg.Username, handle: reg.Handle}

	credential, err := rp.FinishRegistration(identity, session, req)
	if err != nil {
		return nil, nil, errors.Mark(err, ErrUnauthorized)
	}

	user, err := m.GetUser(reg.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user, err = m.CreateUser(reg.Username, WithGroup("user"), withHandle(reg.Handle))
	}
	if err != nil {
		return nil, nil, err
	}

	buf, err := json.Marshal(credential)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	token := &Token{
		Username:     user.Name,
		Kind:         TokenKindPasskey,
		CredentialID: credential.ID,
		Secret: Secret{Passkey: &PasskeySecret{
			Credential: buf,
			Label:      label,
		}},
	}
	if err := m.createToken(token); err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

// BeginPasskeyLogin 은 username 없이 진행하는 로그인 ceremony 를 시작한다.
// 반환된 세션 데이터는 FinishPasskeyLogin 에 그대로 넘겨야 한다.
func (m *Manager) BeginPasskeyLogin(site *url.URL) (*protocol.CredentialAssertion, []byte, error) {
	rp, err := m.relyingParty(site)
	if err != nil {
		return nil, nil, err
	}

	assertion, session, err := rp.BeginDiscoverableLogin()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	buf, err := json.Marshal(session)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return assertion, buf, nil
}

// FinishPasskeyLogin 은 assertion 을 검증하고 로그인한 사용자를 돌려준다.
func (m *Manager) FinishPasskeyLogin(site *url.URL, sessionData []byte, req *http.Request) (*User, error) {
	rp, err := m.relyingParty(site)
	if err != nil {
		return nil, err
	}

	session := webauthn.SessionData{}
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, errors.WithStack(err)
	}

	// 검증에 쓰인 토큰을 밖으로 꺼내 sign counter 를 되돌려 저장한다.
	matched := (*Token)(nil)

	_, credential, err := rp.FinishPasskeyLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
		token, err := m.getPasskeyToken(rawID)
		if err != nil {
			return nil, err
		}

		identity, err := m.identity(token.Username)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(identity.handle, userHandle) {
			return nil, ErrUnauthorized
		}

		matched = token
		return identity, nil
	}, session, req)
	if err != nil {
		return nil, errors.Mark(err, ErrUnauthorized)
	}

	if credential.Authenticator.CloneWarning {
		logrus.Warnf("passkey sign counter went backwards. user: %s", matched.Username)
	}

	buf, err := json.Marshal(credential)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	matched.Secret.Passkey.Credential = buf

	// Update("secret", ...) 는 serializer 를 거치지 않고,
	// Save 는 Index 0 을 새 행으로 봐서 INSERT 가 된다. 구조체 Updates 를 쓴다.
	if err := m.db.Model(&Token{}).
		Where("credential_id = ?", credential.ID).
		Updates(Token{Secret: matched.Secret}).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return m.GetUser(matched.Username)
}

func (m *Manager) getPasskeyToken(credentialID []byte) (*Token, error) {
	token := Token{}
	if err := m.db.Where("kind = ? AND credential_id = ?", TokenKindPasskey, credentialID).
		Take(&token).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &token, nil
}

func (m *Manager) ListPasskey(username string) ([]Token, error) {
	tokens := []Token{}
	if err := m.db.Where(Token{Username: username, Kind: TokenKindPasskey}).
		Order("`index` asc").
		Find(&tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return tokens, nil
}

func (m *Manager) RevokePasskey(username string, index int) error {
	return m.RevokeToken(username, TokenKindPasskey, index)
}

// PasskeyCredential 은 저장된 WebAuthn credential 을 되살린다.
func (token *Token) PasskeyCredential() (*webauthn.Credential, error) {
	if token.Secret.Passkey == nil {
		return nil, errors.Errorf("token %s/%s/%d is not a passkey", token.Username, token.Kind, token.Index)
	}

	credential := webauthn.Credential{}
	if err := json.Unmarshal(token.Secret.Passkey.Credential, &credential); err != nil {
		return nil, errors.WithStack(err)
	}
	return &credential, nil
}
