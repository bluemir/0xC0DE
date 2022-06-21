package auth

import (
	"time"
)

type User struct {
	Name      string `gorm:"primaryKey;size:256"`
	HashedKey string
	Salt      string
}
type Group struct {
	Name string `gorm:"primaryKey;size:256"`
}
type Token struct {
	Name      string `gorm:"primaryKey;size:256"`
	Username  string `gorm:"primaryKey;size:256"`
	HashedKey string `gorm:"primaryKey;size:256"`
	ExpiredAt *time.Time
}
type RoleBinding struct {
	User string
	Role string
}

// not save into db?
type Role struct {
	Name  string `gorm:"primaryKey;size:256"`
	Rules []struct {
		Resource KeyValues
		Verbs    []Verb
	}
}

type Resource interface {
	Get(key string) string
}
type Verb string

type KeyValues map[string]string

func (kvs KeyValues) Get(key string) string {
	return kvs[key]
}
func (kvs KeyValues) isSubsetOf(resource Resource) bool {
	for k, v := range kvs {
		if resource.Get(k) != v {
			return false
		}
	}
	return true
}
