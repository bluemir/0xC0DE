package server

import (
	"github.com/bluemir/0xC0DE/pkg/resources"
	"github.com/gin-gonic/gin"
)

func (server *Server) routes(app gin.IRouter) {
	app.Group("/static", cacheOff).StaticFS("/", resources.Static.HTTPBox())

	app.GET("/", server.static("/index.html"))
}
