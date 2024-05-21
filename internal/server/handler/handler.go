package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	bs "github.com/bluemir/0xC0DE/internal/server/backend"
	backendAuth "github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth"
)

var (
	keyBackends = xid.New().String()
)

func Inject(b *bs.Backends) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(keyBackends, b)
	}
}
func backends(c *gin.Context) *bs.Backends {
	return c.MustGet(keyBackends).(*bs.Backends)
}
func me(c *gin.Context) (*backendAuth.User, error) {
	return auth.User(c)
}

type ListResponse[T any] struct {
	Items []T
}
