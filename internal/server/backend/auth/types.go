package auth

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type User struct {
	Name   string `gorm:"primaryKey;size:256" json:"name" expr:"name"`
	Salt   string `json:"-"`
	Groups Set    `gorm:"type:bytes;serializer:gob" json:"groups"`
	Labels Labels `gorm:"type:bytes;serializer:gob" json:"labels" expr:"labels"`
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
	Rules []Rule `gorm:"type:bytes;serializer:gob"`
}

type Rule struct {
	Resource   KeyValues   `gorm:"type:bytes;serializer:gob"`
	Verbs      []Verb      `gorm:"type:bytes;serializer:gob"`
	Conditions []Condition `gorm:"type:bytes;serializer:gob"`
	//
	// Attribute check?
	// eg)
	// user
	//   subject.kind == "user" && subject.name == object.name
	// project
	//   subject == object.owner
}
type RoleBinding struct {
	Subject   Subject `gorm:"embedded"`
	RoleNames Set     `gorm:"type:bytes;serializer:gob"`
}

type Subject struct {
	Kind string `gorm:"primaryKey;size:256" expr:"kind"`
	Name string `gorm:"primaryKey;size:256" expr:"name"`
}

const (
	KindUser  = "user"
	KindGroup = "group"
	KindGuest = "guest"
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

func (role *Role) IsAllow(ctx Context) bool {
	for _, rule := range role.Rules {
		if rule.IsMatched(ctx) {
			return true
		}
	}
	return false
}
func (rule *Rule) IsMatched(ctx Context) bool {
	if !rule.isVerbMatched(ctx.Verb) {
		return false
	}
	if !rule.Resource.IsSubsetOf(ctx.Resource) {
		return false
	}
	for _, cond := range rule.Conditions {
		if r, err := cond.IsMatched(ctx); err != nil {
			logrus.Error(err)
			return false
		} else if !r {
			return false
		}
	}
	return true
}
func (rule *Rule) isVerbMatched(verb Verb) bool {
	if len(rule.Verbs) == 0 {
		return true
	}
	for _, v := range rule.Verbs {
		if verb == v {
			return true
		}
	}
	return false
}

var x = struct{}{}

func (u *User) Subjects() []Subject {

	if u == nil {
		return []Subject{
			{
				Kind: KindGuest,
			},
		}
	}
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
