package auth

import (
	"encoding/json"
	"time"

	"golang.org/x/exp/maps"
)

const (
	KindUser           = "user"
	KindGroup          = "group"
	KindServiceAccount = "service-account"
	KindGuest          = "guest"
)

type Verb string
type Resource = KeyValues

type KeyValues map[string]string

func (kvs KeyValues) Get(key string) string {
	return kvs[key]
}

func (kvs KeyValues) IsSubsetOf(resource Resource) bool {
	for k, v := range kvs {
		if resource[k] != v {
			return false
		}
	}
	return true
}

type TokenOpt func(*Token)

func ExpiredAt(t time.Time) func(*Token) {
	return func(token *Token) {
		token.ExpiredAt = &t
	}
}
func ExpiredAfter(d time.Duration) func(*Token) {
	return func(token *Token) {
		t := time.Now().Add(d)
		token.ExpiredAt = &t
	}
}

type Set map[string]struct{}

func setFromArray(arr []string) Set {
	s := Set{}
	for _, v := range arr {
		s[v] = struct{}{}
	}
	return s
}

type Labels map[string]string

var x = struct{}{}

func (s Set) Add(vs ...string) {
	for _, v := range vs {
		s[v] = x
	}
}
func (s Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(maps.Keys(s))
}

type Context struct {
	User     *User    `expr:"user"`
	Subject  Subject  `expr:"subject"`
	Verb     Verb     `expr:"verb"`
	Resource Resource `expr:"resource"`
}
