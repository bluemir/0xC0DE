package server

import (
	"github.com/bluemir/0xC0DE/pkg/static"
	"github.com/gin-gonic/gin"
)

func (server *Server) routes(app gin.IRouter) {
	app.Group("/static", staticCache).StaticFS("/", static.Static.HTTPBox())
	app.Group("/lib", staticCache).StaticFS("/", static.NodeModules.HTTPBox()) // for css or other web libs. eg. font-awesome

	app.GET("/", server.static("/index.html"))
}
