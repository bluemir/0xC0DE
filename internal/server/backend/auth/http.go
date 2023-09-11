package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	headerAuthorization   = "Authorization"
	headerWWWAuthenticate = "WWW-Authenticate"
)

func ToHTTPToken(username, unhashedKey string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{username, unhashedKey}, ":")))
}
func ParseHTTPRequest(req *http.Request) (string, string, error) {
	return ParseHTTPHeader(req.Header.Get(headerAuthorization))
}
func (m *Manager) HTTP(req *http.Request) (*User, error) {
	username, key, err := ParseHTTPRequest(req)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return m.Default(username, key)
}
func split2(str string, sep string) (string, string) {
	arr := strings.SplitN(str, sep, 2)
	if len(arr) < 2 {
		return arr[0], ""
	}

	return arr[0], arr[1]
}
func ParseHTTPHeader(header string) (string, string, error) {
	method, data := split2(header, " ")

	switch strings.ToLower(method) {
	case "basic", "token", "bearer":
		c, err := base64.StdEncoding.DecodeString(strings.TrimSpace(data))
		if err != nil {
			return "", "", err
		}

		name, key := split2(string(c), ":")

		return name, key, nil
	default:
		return "", "", ErrUnauthorized // unknown method
	}
}
func LoginHeader(req *http.Request) (string, string) {
	return headerWWWAuthenticate, "basic realm=" + req.URL.Host
}
