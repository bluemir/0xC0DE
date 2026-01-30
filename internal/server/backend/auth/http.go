package auth

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/cockroachdb/errors"
)

const (
	headerAuthorization   = "Authorization"
	headerWWWAuthenticate = "WWW-Authenticate"
)

func (m *Manager) HTTP(req *http.Request) (*User, error) {
	header := req.Header.Get(headerAuthorization)
	scheme, payload := split2(header, " ")

	switch strings.ToLower(scheme) {
	case "basic":
		c, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
		if err != nil {
			return nil, errors.WithStack(err)
		}

		name, password := split2(string(c), ":")

		return m.Default(name, password)
	case "bearer":
		// {username}.{index}.{secret}
		tokenId, secret := split2Last(payload, ".")
		username, idx := split2Last(tokenId, ".")

		index, err := strconv.ParseInt(idx, 10, 32)
		if err != nil {
			return nil, errors.WithStack(err) // maybe 400?
		}

		token, err := m.GetToken(username, TokenKindAccessKey, int(index))
		if err != nil {
			return nil, err
		}

		if err := token.Validate(secret); err != nil {
			return nil, err
		}
		return m.GetUser(username)
	default:
		return nil, meta.ErrNotImplemented
	}
}
func split2(str string, sep string) (string, string) {
	arr := strings.SplitN(str, sep, 2)
	if len(arr) < 2 {
		return arr[0], ""
	}

	return arr[0], arr[1]
}
func split2Last(str, sep string) (string, string) {
	i := strings.LastIndex(str, sep)
	if i < 0 {
		return str, ""
	}

	return str[:i], str[i+len(sep):]
}

func LoginHeader(req *http.Request) (string, string) {
	return headerWWWAuthenticate, "basic realm=" + req.URL.Host
}
