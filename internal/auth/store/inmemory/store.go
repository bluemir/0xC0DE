package inmemory

import "github.com/bluemir/0xC0DE/internal/auth"

func New(salt string) (auth.AuthStore, error) {
	return &Store{
		users:        map[string]auth.User{},
		tokens:       map[TokenIndex]auth.Token{},
		groups:       map[string]auth.Group{},
		roles:        map[string]auth.Role{},
		rolebindings: map[auth.Subject]auth.RoleBinding{},
		salt:         salt,
	}, nil
}

type Store struct {
	users        map[string]auth.User
	tokens       map[TokenIndex]auth.Token
	groups       map[string]auth.Group
	roles        map[string]auth.Role
	rolebindings map[auth.Subject]auth.RoleBinding
	salt         string
}
type TokenIndex struct {
	Username  string
	HashedKey string
}
