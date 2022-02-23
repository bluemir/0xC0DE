package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Success 200 {object} object{message=string}
// @Router /api/v1/ping [get]
func (server *Server) ping(c *gin.Context) {
	// for example
	if err := server.db.AutoMigrate(); err != nil {
		APIErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
