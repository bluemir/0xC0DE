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
func (server *Server) routes(app gin.IRouter, noRoute func(...gin.HandlerFunc)) {
	// prometheus for monitoring
	app.GET("/metric", prom.Handler())
	app.Use(prom.Metrics())

	// swagger
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// API
	{
		v1 := app.Group("/api/v1", markAPI)

		v1.GET("/ping", handler.Ping)
		v1.GET("/authn/ping", auth.RequireLogin, handler.Ping)
		v1.GET("/authz/ping", auth.Can(verb.Create, resource.Server), handler.Ping)
		// roles:
		// - name: admin
		//   rules:
		//   - resource:
		//       kind: foo
		//       name: bar
		//     verb: create

		v1.POST("/login", auth.Login)
		v1.GET("/logout", auth.Logout)
		v1.POST("/users", handler.Register)
		v1.GET("/users/me", handler.Me)

		v1.GET("/users", handler.ListUsers)
		v1.GET("/groups", handler.ListGroups)
		v1.GET("/roles", handler.ListRoles)

		v1.POST("/posts", handler.CreatePost)
		v1.GET("/posts", handler.ListPost)
		v1.GET("/posts/stream", handler.StreamPost)

		// WebSocket
		v1.GET("/ws", handler.Websocket)
		// Server Sent Event
		v1.GET("/stream", handler.Stream)
		// http2 Server Push
		v1.GET("/push", handler.Push)
	}

	// Static Pages
	{
		// js, css, etc.
		app.Group("/static", staticCache()).StaticFS("/", http.FS(static.Static))

		app.GET("/", html("index.html"))
		app.GET("/users/register", html("register.html"))
		app.GET("/users/login", html("login.html"))
		app.GET("/posts", html("posts.html"))
		app.GET("/admin", html("admin.html"))
		app.GET("/admin/iam", redirect("/admin/iam/users"))
		app.GET("/admin/iam/users", html("admin/iam/users.html"))
		app.GET("/admin/iam/groups", html("admin/iam/groups.html"))
		app.GET("/admin/iam/roles", html("admin/iam/roles.html"))

		// or for SPA(single page application), client side routing
		// app.Use(AbortIfHasPrefix("/api"), server.static("/index.html"))
	}
}
