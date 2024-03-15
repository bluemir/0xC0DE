package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/middleware/auth"
)

func Me(c *gin.Context) {
	u, err := auth.User(c)

	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, u)
}
