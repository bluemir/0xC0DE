package injector

import (
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/mail"
	"github.com/bluemir/0xC0DE/internal/server/backend"
)

type keyBackend struct{}
type keyMail struct{}

func Inject(backends *backend.Backends, mailSender *mail.Sender) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keyBackend{}, backends)
		c.Set(keyMail{}, mailSender)
	}
}
func Backends(c *gin.Context) *backend.Backends {
	return c.MustGet(keyBackend{}).(*backend.Backends)
}
func Mail(c *gin.Context) *mail.Sender {
	return c.MustGet(keyMail{}).(*mail.Sender)
}
