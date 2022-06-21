package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

const (
	httpAuthorizationHeader = "Authorization"
)

func (m *Manager) HTTP(req *http.Request) (*User, error) {
	username, key, err := parseHTTPHeader(req.Header.Get(httpAuthorizationHeader))
	if err != nil {
		return nil, err
	}

	return m.Default(username, key)
}
func (m *Manager) NewHTTPToken(username string, expireAt time.Time) (string, error) {
	user := &User{}
	if err := m.db.Where(&User{Name: username}).Take(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", ErrUnauthroized
		}
		return "", err
	}

	newKey := hash(xid.New().String(), user.Salt)

	if err := m.IssueToken(username, newKey, &expireAt); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{user.Name, newKey}, ":"))), nil
}
func (m *Manager) RevokeHTTPToken(req *http.Request) error {
	username, key, err := parseHTTPHeader(req.Header.Get(httpAuthorizationHeader))
	if err != nil {

		return err
	}
	return m.RevokeToken(username, key)
}

func split2(str string, sep string) (string, string) {
	arr := strings.SplitN(str, sep, 2)
	if len(arr) < 2 {
		return arr[0], ""
	}

	return arr[0], arr[1]
}
func parseHTTPHeader(header string) (string, string, error) {
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
		return "", "", ErrUnauthroized // unknown method
	}
}
