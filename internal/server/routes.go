package server

import (
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/middleware/auth"
	"github.com/bluemir/0xC0DE/internal/server/middleware/prom"
	"github.com/bluemir/0xC0DE/internal/static"

	// swagger
	_ "github.com/bluemir/0xC0DE/internal/swagger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	VerbCreate auth.Verb = "CREATE"
)

func ResourcePing(c *gin.Context) auth.Resource {
	return auth.KeyValues{"kind": "ping"}
}

// @title 0xC0DE
// @version 0.1.0
// @description
func (server *Server) routes(app gin.IRouter) {
	// prometheus for monitoring
	app.GET("/metric", prom.Handler())
	app.Use(prom.Metrics())

	// swagger
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// API
	{
		v1 := app.Group("/api/v1", markAPI)

		v1.GET("/ping", server.handler.Ping)

		v1.POST("/token", auth.IssueToken)
		v1.GET("/authn/ping", auth.RequireLogin, server.handler.Ping)
		v1.GET("/authz/ping", auth.Authz(ResourcePing, VerbCreate), server.handler.Ping)
		// roles:
		// - name: admin
		//   rules:
		//   - resource:
		//       kind: foo
		//       name: bar
		//     verb: create

	}

	// WebSocket
	app.GET("/ws", server.websocket)

	// Static Pages
	{
		// js, css, etc.
		app.Group("/static", server.staticCache).StaticFS("/", static.Static.HTTPBox())
		//app.Group("/lib", server.staticCache).StaticFS("/", static.NodeModules.HTTPBox()) // for css or other web libs. eg. font-awesome

		app.GET("/", markHTML, server.static("/index.html"))
		// or for SPA(single page application), client side routing
		// app.Use(AbortIfHasPrefix("/api"), server.static("/index.html"))
	}
}
