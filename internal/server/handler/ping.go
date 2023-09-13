package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Success 200 {object} object{message=string}
// @Router /api/v1/ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
