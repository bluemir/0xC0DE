package composite

import "github.com/bluemir/0xC0DE/internal/server/backend/auth"

type Store struct {
	auth.AuthUserStore
	auth.AuthTokenStore
	auth.AuthGroupStore
	auth.AuthRoleStore
	auth.AuthRoleBindingStore
}
