package server

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
)

func html(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		markHTML(c)
		c.HTML(http.StatusOK, path, c)
	}
}
func redirect(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, path)
	}
}

func (server *Server) AbortIfHasPrefix(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, prefix) {
			c.Abort()
		}
	}
}

func fixURL(c *gin.Context) {
	url := location.Get(c)

	// QUESTION is it right?
	c.Request.URL.Scheme = url.Scheme
	c.Request.URL.Host = url.Host
}
func markAPI(c *gin.Context) {
	c.SetAccepted("application/json")
}
func markHTML(c *gin.Context) {
	c.SetAccepted("text/html")
}
