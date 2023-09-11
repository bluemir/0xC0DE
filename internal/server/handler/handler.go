package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/bus"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func New(db *gorm.DB) (*Handler, error) {
	return &Handler{
		db: db,
	}, nil
}

type Handler struct {
	db *gorm.DB
}

var (
	keyBackends = xid.New().String()
)

func Inject(b *Backends) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set(keyBackends, b)
	}
}
func backends(c *gin.Context) *Backends {
	return c.MustGet(keyBackends).(*Backends)
}

type Backends struct {
	Auth     *auth.Manager
	EventBus *bus.Bus
}
