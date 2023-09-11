package resource

import (
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func Server(c *gin.Context) auth.Resource {
	return auth.KeyValues{
		"kind": "server",
	}
}
func User(c *gin.Context) auth.Resource {
	username := c.Param("username")

	return auth.KeyValues{
		"kind": "user",
		"name": username,
	}
}
