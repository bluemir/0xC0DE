package injector

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	"github.com/bluemir/0xC0DE/internal/server/backend"
)

var keyBackend = xid.New().String()

func Inject(b *backend.Backends) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keyBackend, b)
	}
}
func Backends(c *gin.Context) *backend.Backends {
	return c.MustGet(keyBackend).(*backend.Backends)
}
