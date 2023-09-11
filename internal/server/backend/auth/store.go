package auth

type AuthStore interface {
	AuthUserStore
	AuthTokenStore
	AuthGroupStore
	AuthRoleStore
	AuthRoleBindingStore
}

type AuthUserStore interface {
	// User CRUD
	CreateUser(*User) error
	GetUser(username string) (*User, error)
	ListUser() ([]User, error)
	UpdateUser(user *User) error
	DeleteUser(username string) error
}
type AuthTokenStore interface {
	// Token CRD
	CreateToken(*Token) error
	GetToken(username, hashedKey string) (*Token, error)
	ListToken(username string) ([]Token, error)
	DeleteToken(username, revokeKey string) error
}
type AuthGroupStore interface {
	// Group CRD
	CreateGroup(*Group) error
	GetGroup(name string) (*Group, error)
	ListGroup() ([]Group, error)
	DeleteGroup(name string) error
}
type AuthRoleStore interface {
	// Role CRUD
	CreateRole(*Role) error
	GetRole(name string) (*Role, error)
	ListRole() ([]Role, error)
	UpdateRole(role *Role) error
	DeleteRole(name string) error
}
type AuthRoleBindingStore interface {
	// RoleBinding CRD
	CreateRoleBinding(*RoleBinding) error
	GetRoleBinding(Subject) (*RoleBinding, error)
	ListRoleBinding() ([]RoleBinding, error)
	UpdateRoleBinding(*RoleBinding) error
	DeleteRoleBinding(Subject) error
}
