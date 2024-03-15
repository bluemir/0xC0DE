package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	bs "github.com/bluemir/0xC0DE/internal/server/backend"
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

type ListResponse[T any] struct {
	Items []T
}
