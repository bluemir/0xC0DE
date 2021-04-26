package server

import (
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/static"
)

func (server *Server) routes(app gin.IRouter) {
	// js, css, etc.
	app.Group("/static", staticCache).StaticFS("/", static.Static.HTTPBox())
	app.Group("/lib", staticCache).StaticFS("/", static.NodeModules.HTTPBox()) // for css or other web libs. eg. font-awesome

	// Static Pages
	{
		app.GET("/", server.static("/index.html"))
	}

	// API
	{
		v1 := app.Group("/api/v1")
		v1.GET("/ping", server.ping)
		v1.GET("authed-ping", server.authAPI, server.ping)
	}

	// WebSocket
	app.GET("/ws", server.websocket)
}
