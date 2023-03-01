package composite

import "github.com/bluemir/0xC0DE/internal/auth"

type Store struct {
	auth.AuthUserStore
	auth.AuthTokenStore
	auth.AuthGroupStore
	auth.AuthRoleStore
	auth.AuthRoleBindingStore
}
