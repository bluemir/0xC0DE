package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Me(c *gin.Context) {
	u, err := me(c)

	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, u)
}
