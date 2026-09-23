package injector

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	"github.com/bluemir/0xC0DE/internal/mail"
	"github.com/bluemir/0xC0DE/internal/server/backend"
)

var (
	keyBackend = xid.New().String()
	keyMail    = xid.New().String()
)

func Inject(b *backend.Backends, m *mail.Sender) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keyBackend, b)
		c.Set(keyMail, m)
	}
}
func Backends(c *gin.Context) *backend.Backends {
	return c.MustGet(keyBackend).(*backend.Backends)
}
func Mail(c *gin.Context) *mail.Sender {
	return c.MustGet(keyMail).(*mail.Sender)
}
