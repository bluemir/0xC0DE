package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

// authenticator 는 테스트용 소프트웨어 authenticator 다.
// 실제 기기 없이 등록/로그인 ceremony 응답을 만든다.
type authenticator struct {
	key          *ecdsa.PrivateKey
	credentialID []byte
	signCount    uint32
}

func newAuthenticator(t *testing.T) *authenticator {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	credentialID := make([]byte, 16)
	_, err = rand.Read(credentialID)
	require.NoError(t, err)

	return &authenticator{key: key, credentialID: credentialID}
}

const (
	flagUserPresent    = 0x01
	flagUserVerified   = 0x04
	flagBackupEligible = 0x08
	flagBackupState    = 0x10
	flagAttestedData   = 0x40
)

func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func (a *authenticator) clientData(t *testing.T, ceremony, challenge, origin string) []byte {
	buf, err := json.Marshal(map[string]any{
		"type":        ceremony,
		"challenge":   challenge,
		"origin":      origin,
		"crossOrigin": false,
	})
	require.NoError(t, err)

	return buf
}

func (a *authenticator) authData(t *testing.T, rpID string, flags byte, attested bool) []byte {
	rpIDHash := sha256.Sum256([]byte(rpID))

	data := []byte{}
	data = append(data, rpIDHash[:]...)
	data = append(data, flags)
	data = binary.BigEndian.AppendUint32(data, a.signCount)

	if !attested {
		return data
	}

	// COSE_Key (ES256)
	coseKey, err := cbor.Marshal(map[int]any{
		1:  2,                                             // kty: EC2
		3:  -7,                                            // alg: ES256
		-1: 1,                                             // crv: P-256
		-2: a.key.PublicKey.X.FillBytes(make([]byte, 32)), // x
		-3: a.key.PublicKey.Y.FillBytes(make([]byte, 32)), // y
	})
	require.NoError(t, err)

	data = append(data, make([]byte, 16)...) // aaguid
	data = binary.BigEndian.AppendUint16(data, uint16(len(a.credentialID)))
	data = append(data, a.credentialID...)
	data = append(data, coseKey...)

	return data
}

// create 는 navigator.credentials.create() 의 응답을 만든다.
func (a *authenticator) create(t *testing.T, rpID, origin, challenge string) *protocol.ParsedCredentialCreationData {
	clientData := a.clientData(t, "webauthn.create", challenge, origin)
	authData := a.authData(t, rpID, flagUserPresent|flagUserVerified|flagBackupEligible|flagBackupState|flagAttestedData, true)

	attestation, err := cbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	require.NoError(t, err)

	parsed, err := protocol.ParseCredentialCreationResponseBytes(jsonBody(t, map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"attestationObject": b64(attestation),
		},
	}))
	require.NoError(t, err)

	return parsed
}

// get 는 navigator.credentials.get() 의 응답을 만든다.
func (a *authenticator) get(t *testing.T, rpID, origin, challenge string, handle []byte) *protocol.ParsedCredentialAssertionData {
	a.signCount++

	clientData := a.clientData(t, "webauthn.get", challenge, origin)
	authData := a.authData(t, rpID, flagUserPresent|flagUserVerified|flagBackupEligible|flagBackupState, false)

	clientDataHash := sha256.Sum256(clientData)
	signed := sha256.Sum256(append(authData, clientDataHash[:]...))

	signature, err := ecdsa.SignASN1(rand.Reader, a.key, signed[:])
	require.NoError(t, err)

	parsed, err := protocol.ParseCredentialRequestResponseBytes(jsonBody(t, map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"authenticatorData": b64(authData),
			"signature":         b64(signature),
			"userHandle":        b64(handle),
		},
	}))
	require.NoError(t, err)

	return parsed
}

func jsonBody(t *testing.T, body map[string]any) []byte {
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	return buf
}

func TestPasskeyRegisterAndLogin(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	const (
		rpID   = "localhost"
		origin = "http://localhost:8080"
	)

	device := newAuthenticator(t)

	// passkey 만으로 가입
	creation, registration, err := m.BeginPasskeyRegistration(ctx, site, "alice")
	require.NoError(t, err)
	assert.Equal(t, rpID, creation.Response.RelyingParty.ID)
	assert.Equal(t, "alice", registration.Username)

	user, token, err := m.FinishPasskeyRegistration(ctx, site, registration, "my laptop",
		device.create(t, rpID, origin, creation.Response.Challenge.String()))
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
	assert.Equal(t, "my laptop", token.Secret.Passkey.Label)

	// password 는 없으므로 ID/PW 로그인은 실패해야 한다
	_, err = m.Default(ctx, "alice", "")
	assert.ErrorIs(t, err, auth.ErrUnauthorized)

	// username 없이 passkey 로 로그인
	assertion, sessionData, err := m.BeginPasskeyLogin(ctx, site)
	require.NoError(t, err)

	loggedIn, err := m.FinishPasskeyLogin(ctx, site, sessionData,
		device.get(t, rpID, origin, assertion.Response.Challenge.String(), user.Handle))
	require.NoError(t, err)
	assert.Equal(t, "alice", loggedIn.Name)

	// 등록된 passkey 목록
	passkeys, err := m.ListPasskey(ctx, "alice")
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	assert.Equal(t, device.credentialID, passkeys[0].ExternalID)

	// sign counter 가 저장되어야 한다
	credential, err := passkeys[0].PasskeyCredential()
	require.NoError(t, err)
	assert.Equal(t, device.signCount, credential.Authenticator.SignCount)
	assert.False(t, credential.Authenticator.CloneWarning)

	require.NoError(t, m.RevokePasskey(ctx, "alice", passkeys[0].Index))

	passkeys, err = m.ListPasskey(ctx, "alice")
	require.NoError(t, err)
	assert.Len(t, passkeys, 0)
}

func TestPasskeyLoginWithUnknownCredential(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	device := newAuthenticator(t)

	assertion, sessionData, err := m.BeginPasskeyLogin(ctx, site)
	require.NoError(t, err)

	_, err = m.FinishPasskeyLogin(ctx, site, sessionData,
		device.get(t, "localhost", "http://localhost:8080", assertion.Response.Challenge.String(), []byte("nobody")))
	// 표시해 두지 않으면 errs 미들웨어가 401 대신 500 을 낸다
	assert.ErrorIs(t, err, auth.ErrUnauthorized)
}

// 등록 ceremony 가 어긋나도 401 이어야 한다
func TestPasskeyRegisterWithWrongChallenge(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	device := newAuthenticator(t)

	_, registration, err := m.BeginPasskeyRegistration(ctx, site, "alice")
	require.NoError(t, err)

	// 서버가 낸 것이 아닌 challenge 로 응답을 만든다
	_, _, err = m.FinishPasskeyRegistration(ctx, site, registration, "my laptop",
		device.create(t, "localhost", "http://localhost:8080", b64([]byte("not-the-challenge"))))
	assert.ErrorIs(t, err, auth.ErrUnauthorized)

	// 실패했으므로 계정도 생기지 않는다
	_, err = m.GetUser(ctx, "alice")
	assert.Error(t, err)
}

func TestPasskeyRegisterForExistingPasswordUser(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	_, _, err = m.Register(ctx, "bob", "password")
	require.NoError(t, err)

	device := newAuthenticator(t)

	creation, registration, err := m.BeginPasskeyRegistration(ctx, site, "bob")
	require.NoError(t, err)

	_, _, err = m.FinishPasskeyRegistration(ctx, site, registration, "yubikey",
		device.create(t, "localhost", "http://localhost:8080", creation.Response.Challenge.String()))
	require.NoError(t, err)

	// ID/PW 로그인은 그대로 동작해야 한다
	_, err = m.Default(ctx, "bob", "password")
	require.NoError(t, err)

	// passkey 로도 로그인된다
	assertion, sessionData, err := m.BeginPasskeyLogin(ctx, site)
	require.NoError(t, err)

	user, err := m.GetUser(ctx, "bob")
	require.NoError(t, err)

	loggedIn, err := m.FinishPasskeyLogin(ctx, site, sessionData,
		device.get(t, "localhost", "http://localhost:8080", assertion.Response.Challenge.String(), user.Handle))
	require.NoError(t, err)
	assert.Equal(t, "bob", loggedIn.Name)
}

func TestPasskeyLastUsedAt(t *testing.T) {
	ctx := t.Context()
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	const (
		rpID   = "localhost"
		origin = "http://localhost:8080"
	)

	device := newAuthenticator(t)

	creation, registration, err := m.BeginPasskeyRegistration(ctx, site, "carol")
	require.NoError(t, err)

	user, _, err := m.FinishPasskeyRegistration(ctx, site, registration, "my laptop",
		device.create(t, rpID, origin, creation.Response.Challenge.String()))
	require.NoError(t, err)

	// 등록만 해서는 아직 로그인에 쓰인 적이 없다
	passkeys, err := m.ListPasskey(ctx, "carol")
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	assert.Nil(t, passkeys[0].Secret.Passkey.LastUsedAt)

	login := func() {
		assertion, sessionData, err := m.BeginPasskeyLogin(ctx, site)
		require.NoError(t, err)

		_, err = m.FinishPasskeyLogin(ctx, site, sessionData,
			device.get(t, rpID, origin, assertion.Response.Challenge.String(), user.Handle))
		require.NoError(t, err)
	}

	login()

	passkeys, err = m.ListPasskey(ctx, "carol")
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	require.NotNil(t, passkeys[0].Secret.Passkey.LastUsedAt)
	assert.WithinDuration(t, time.Now(), *passkeys[0].Secret.Passkey.LastUsedAt, time.Minute)

	// 다음 로그인이 값을 덮는지 본다.
	// 두 번을 연달아 부르면 시각이 같게 나올 수 있어 저장된 값을 과거로 돌려 놓는다.
	stale := time.Now().Add(-time.Hour)
	staleSecret := passkeys[0].Secret
	staleSecret.Passkey.LastUsedAt = &stale

	require.NoError(t, db.Model(&auth.Token{}).
		Where("kind = ? AND external_id = ?", auth.TokenKindPasskey, passkeys[0].ExternalID).
		Updates(auth.Token{Secret: staleSecret}).Error)

	login()

	passkeys, err = m.ListPasskey(ctx, "carol")
	require.NoError(t, err)
	require.Len(t, passkeys, 1)

	lastUsedAt := passkeys[0].Secret.Passkey.LastUsedAt
	require.NotNil(t, lastUsedAt)
	assert.True(t, lastUsedAt.After(stale), "last used at should move forward: %s", lastUsedAt)
}

// sign counter 가 뒤로 가도 로그인은 되고, 경고는 역행 한 번에 한 줄만 남아야 한다.
func TestPasskeyLoginWithRegressedSignCounter(t *testing.T) {
	ctx := t.Context()
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	const (
		rpID   = "localhost"
		origin = "http://localhost:8080"
	)

	device := newAuthenticator(t)

	creation, registration, err := m.BeginPasskeyRegistration(ctx, site, "dave")
	require.NoError(t, err)

	user, _, err := m.FinishPasskeyRegistration(ctx, site, registration, "yubikey",
		device.create(t, rpID, origin, creation.Response.Challenge.String()))
	require.NoError(t, err)

	login := func() {
		assertion, sessionData, err := m.BeginPasskeyLogin(ctx, site)
		require.NoError(t, err)

		loggedIn, err := m.FinishPasskeyLogin(ctx, site, sessionData,
			device.get(t, rpID, origin, assertion.Response.Challenge.String(), user.Handle))
		require.NoError(t, err)
		assert.Equal(t, "dave", loggedIn.Name)
	}

	storedCounter := func() uint32 {
		passkeys, err := m.ListPasskey(ctx, "dave")
		require.NoError(t, err)
		require.Len(t, passkeys, 1)

		credential, err := passkeys[0].PasskeyCredential()
		require.NoError(t, err)
		assert.False(t, credential.Authenticator.CloneWarning, "플래그를 저장하면 다음 로그인에서 또 경고가 난다")

		return credential.Authenticator.SignCount
	}

	device.signCount = 10
	login()
	require.Equal(t, uint32(11), storedCounter())

	// 키를 초기화해 카운터가 0 으로 돌아간 것처럼 만든다
	device.signCount = 0

	hook := test.NewGlobal()

	login()
	assert.Equal(t, uint32(1), storedCounter(), "받은 값을 새 기준으로 삼아야 한다")
	// 테스트는 TraceLevel 이라 레벨을 짚지 않으면 Debug 로 내려가도 통과한다
	require.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "sign counter went backwards")

	// 이어지는 로그인은 조용해야 한다
	hook.Reset()

	login()
	assert.Equal(t, uint32(2), storedCounter())
	assert.Empty(t, hook.AllEntries())
}
