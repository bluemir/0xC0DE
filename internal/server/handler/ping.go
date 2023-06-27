package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Success 200 {object} object{message=string}
// @Router /api/v1/ping [get]
func (h *Handler) Ping(c *gin.Context) {
	// for example
	if err := h.db.AutoMigrate(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
