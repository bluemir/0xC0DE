package gorm

import "github.com/bluemir/0xC0DE/internal/auth"

type User = auth.User
type Token = auth.Token
type Group = auth.Group
type Role = auth.Role
type Resource = auth.Resource
type Verb = auth.Verb

type RoleBinding struct {
	User string
	Role string
}
