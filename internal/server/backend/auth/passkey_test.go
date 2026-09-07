package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
func (a *authenticator) create(t *testing.T, rpID, origin, challenge string) *http.Request {
	clientData := a.clientData(t, "webauthn.create", challenge, origin)
	authData := a.authData(t, rpID, flagUserPresent|flagUserVerified|flagBackupEligible|flagBackupState|flagAttestedData, true)

	attestation, err := cbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	require.NoError(t, err)

	return jsonRequest(t, map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"attestationObject": b64(attestation),
		},
	})
}

// get 는 navigator.credentials.get() 의 응답을 만든다.
func (a *authenticator) get(t *testing.T, rpID, origin, challenge string, handle []byte) *http.Request {
	a.signCount++

	clientData := a.clientData(t, "webauthn.get", challenge, origin)
	authData := a.authData(t, rpID, flagUserPresent|flagUserVerified|flagBackupEligible|flagBackupState, false)

	clientDataHash := sha256.Sum256(clientData)
	signed := sha256.Sum256(append(authData, clientDataHash[:]...))

	signature, err := ecdsa.SignASN1(rand.Reader, a.key, signed[:])
	require.NoError(t, err)

	return jsonRequest(t, map[string]any{
		"id":    b64(a.credentialID),
		"rawId": b64(a.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"authenticatorData": b64(authData),
			"signature":         b64(signature),
			"userHandle":        b64(handle),
		},
	})
}

func jsonRequest(t *testing.T, body map[string]any) *http.Request {
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(buf)))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func TestPasskeyRegisterAndLogin(t *testing.T) {
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
	creation, registration, err := m.BeginPasskeyRegistration(site, "alice")
	require.NoError(t, err)
	assert.Equal(t, rpID, creation.Response.RelyingParty.ID)
	assert.Equal(t, "alice", registration.Username)

	user, token, err := m.FinishPasskeyRegistration(site, registration, "my laptop",
		device.create(t, rpID, origin, creation.Response.Challenge.String()))
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
	assert.Equal(t, "my laptop", token.Label)

	// password 는 없으므로 ID/PW 로그인은 실패해야 한다
	_, err = m.Default("alice", "")
	assert.Error(t, err)

	// username 없이 passkey 로 로그인
	assertion, sessionData, err := m.BeginPasskeyLogin(site)
	require.NoError(t, err)

	loggedIn, err := m.FinishPasskeyLogin(site, sessionData,
		device.get(t, rpID, origin, assertion.Response.Challenge.String(), user.Handle))
	require.NoError(t, err)
	assert.Equal(t, "alice", loggedIn.Name)

	// 등록된 passkey 목록
	passkeys, err := m.ListPasskey("alice")
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	assert.Equal(t, device.credentialID, passkeys[0].CredentialID)

	// sign counter 가 저장되어야 한다
	credential, err := passkeys[0].PasskeyCredential()
	require.NoError(t, err)
	assert.Equal(t, device.signCount, credential.Authenticator.SignCount)
	assert.False(t, credential.Authenticator.CloneWarning)

	require.NoError(t, m.RevokePasskey("alice", passkeys[0].Index))

	passkeys, err = m.ListPasskey("alice")
	require.NoError(t, err)
	assert.Len(t, passkeys, 0)
}

func TestPasskeyLoginWithUnknownCredential(t *testing.T) {
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	device := newAuthenticator(t)

	assertion, sessionData, err := m.BeginPasskeyLogin(site)
	require.NoError(t, err)

	_, err = m.FinishPasskeyLogin(site, sessionData,
		device.get(t, "localhost", "http://localhost:8080", assertion.Response.Challenge.String(), []byte("nobody")))
	assert.Error(t, err)
}

func TestPasskeyRegisterForExistingPasswordUser(t *testing.T) {
	m, err := newManager()
	require.NoError(t, err)

	site, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	_, _, err = m.Register("bob", "password")
	require.NoError(t, err)

	device := newAuthenticator(t)

	creation, registration, err := m.BeginPasskeyRegistration(site, "bob")
	require.NoError(t, err)

	_, _, err = m.FinishPasskeyRegistration(site, registration, "yubikey",
		device.create(t, "localhost", "http://localhost:8080", creation.Response.Challenge.String()))
	require.NoError(t, err)

	// ID/PW 로그인은 그대로 동작해야 한다
	_, err = m.Default("bob", "password")
	require.NoError(t, err)

	// passkey 로도 로그인된다
	assertion, sessionData, err := m.BeginPasskeyLogin(site)
	require.NoError(t, err)

	user, err := m.GetUser("bob")
	require.NoError(t, err)

	loggedIn, err := m.FinishPasskeyLogin(site, sessionData,
		device.get(t, "localhost", "http://localhost:8080", assertion.Response.Challenge.String(), user.Handle))
	require.NoError(t, err)
	assert.Equal(t, "bob", loggedIn.Name)
}
