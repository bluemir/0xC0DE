package server

import (
	"github.com/bluemir/0xC0DE/pkg/static"
	"github.com/gin-gonic/gin"
)

func (server *Server) routes(app gin.IRouter) {
	app.Group("/static", cacheOff).StaticFS("/", static.Static.HTTPBox())

	app.GET("/", server.static("/index.html"))
}
