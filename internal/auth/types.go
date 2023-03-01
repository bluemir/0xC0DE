package auth

import (
	"time"
)

type User struct {
	Name   string `gorm:"primaryKey;size:256"`
	Salt   string
	Groups Set `sql:"type:json"`
	Labels Labels
}
type Group struct {
	Name string `gorm:"primaryKey;size:256"`
}
type Token struct {
	Username  string `gorm:"primaryKey;size:256"`
	HashedKey string `gorm:"primaryKey;size:256"`
	RevokeKey string
	ExpiredAt *time.Time
}

type Role struct {
	Name  string `gorm:"primaryKey;size:256"`
	Rules []Rule
}

type Rule struct {
	Resource KeyValues
	Verbs    []Verb
}
type RoleBinding struct {
	Subject   Subject
	RoleNames Set
}

type Subject struct {
	Kind string `gorm:"primaryKey;size:256"`
	Name string `gorm:"primaryKey;size:256"`
}

const (
	KindUser  = "user"
	KindGroup = "group"
)

type Resource interface {
	Get(key string) string
}

type Verb string

type KeyValues map[string]string

func (kvs KeyValues) Get(key string) string {
	return kvs[key]
}
func (kvs KeyValues) IsSubsetOf(resource Resource) bool {
	for k, v := range kvs {
		if resource.Get(k) != v {
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

type Labels map[string]string

func (role *Role) IsAllow(verb Verb, resource Resource) bool {
	for _, rule := range role.Rules {
		for _, v := range rule.Verbs {
			if rule.Resource.IsSubsetOf(resource) && v == verb {
				return true
			}
		}
	}
	return false
}

var x = struct{}{}

func (u *User) Subjects() []Subject {
	ret := []Subject{
		{
			Kind: KindUser,
			Name: u.Name,
		},
	}
	for g := range u.Groups {
		ret = append(ret, Subject{
			Kind: KindGroup,
			Name: g,
		})
	}
	return ret
}
