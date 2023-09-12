package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/handler"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth/resource"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth/verb"
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

	// API
	{
		v1 := app.Group("/api/v1", markAPI)

		v1.GET("/ping", server.handler.Ping)

		v1.GET("/login", auth.Login)
		v1.GET("/logout", auth.Logout)
		v1.POST("/users", handler.Register)
		v1.GET("/authn/ping", auth.RequireLogin, server.handler.Ping)
		v1.GET("/authz/ping", auth.Can(verb.Create, resource.Server), server.handler.Ping)
		// roles:
		// - name: admin
		//   rules:
		//   - resource:
		//       kind: foo
		//       name: bar
		//     verb: create
	}

	// WebSocket
	app.GET("/ws", server.handler.Websocket)
	// Server Sent Event
	app.GET("/stream", server.handler.Stream)
	// http2 Server Push
	app.GET("/push", server.handler.Push)

	// Static Pages
	{
		// js, css, etc.
		app.Group("/static", staticCache()).StaticFS("/", http.FS(static.Static))

		app.GET("/", HTML("index.html"))
		app.GET("/posts", HTML("posts.html"))
		app.GET("/admin", HTML("admin.html"))
		app.GET("/admin/users", HTML("admin/users.html"))
		// or for SPA(single page application), client side routing
		// app.Use(AbortIfHasPrefix("/api"), server.static("/index.html"))
	}
}
