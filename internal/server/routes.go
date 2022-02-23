package server

import (
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/middleware/prom"
	"github.com/bluemir/0xC0DE/internal/static"

	// swagger
	_ "github.com/bluemir/0xC0DE/internal/swagger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 0xC0DE
// @version 0.1.0
// @description
func (server *Server) routes(app gin.IRouter) {
	// prometheus for monitoring
	app.GET("/metric", prom.Handler())
	app.Use(prom.Metrics())

	// swagger
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// js, css, etc.
	app.Group("/static", server.staticCache).StaticFS("/", static.Static.HTTPBox())
	//app.Group("/lib", server.staticCache).StaticFS("/", static.NodeModules.HTTPBox()) // for css or other web libs. eg. font-awesome

	// API
	{
		v1 := app.Group("/api/v1")
		v1.GET("/ping", server.ping)
		v1.GET("authed-ping", server.authAPI, server.ping)
	}

	// WebSocket
	app.GET("/ws", server.websocket)

	// Static Pages
	{
		app.GET("/", server.static("/index.html"))
		// or for SPA(single page application), client side routing
		// app.Use(AbortIfHasPrefix("/api"), server.static("/index.html"))
	}
}
