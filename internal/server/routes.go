package server

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/assets"
	"github.com/bluemir/0xC0DE/internal/server/handler"
	"github.com/bluemir/0xC0DE/internal/server/handler/auth/resource"
	"github.com/bluemir/0xC0DE/internal/server/handler/auth/verb"
	"github.com/bluemir/0xC0DE/internal/server/middleware/bootstrap"
	"github.com/bluemir/0xC0DE/internal/server/middleware/cache"
)

// @title 0xC0DE
// @version 0.1.0
// @description
func (server *Server) routes(app gin.IRouter, noRoute func(...gin.HandlerFunc)) {
	var (
		requireLogin = handler.RequireLogin
		can          = handler.Can
	)

	// API
	{
		v1 := app.Group("/api/v1", markAcceptJSON)

		v1.GET("/ping", api(handler.Ping))
		v1.GET("/authn/ping", requireLogin, api(handler.Ping))
		v1.GET("/authz/ping", can(verb.Create, resource.Server), api(handler.Ping))
		// roles:
		// - name: admin
		//   rules:
		//   - resource:
		//       kind: foo
		//       name: bar
		//     verb: create

		v1.POST("/login", api(handler.Login))
		v1.GET("/logout", api(handler.Logout))
		v1.POST("/users", api(handler.Register))
		v1.GET("/users/me", api(handler.Me))

		v1.POST("/bootstrap", bodyReaderTweak, bootstrap.CheckBootstrapToken, api(handler.Register))

		v1.GET("/can/:verb/:resource", api(handler.CanAPI))
		v1.GET("/can/:verb", api(handler.CanAPI))
		v1.GET("/can", api(handler.CanBulkAPI))

		v1.GET("/users", api(handler.ListUsers))
		v1.GET("/groups", api(handler.ListGroups))
		v1.GET("/roles", api(handler.ListRoles))

		v1.POST("/posts", api(handler.CreatePost))
		v1.GET("/posts", api(handler.FindPost))
		v1.GET("/posts/stream", sse(handler.StreamPost))

		// WebSocket
		v1.GET("/ws", handler.Websocket)
		// Server Sent Event
		v1.GET("/stream", sse(handler.Stream))
		// http2 Server Push
		v1.GET("/push", api(handler.Push))
	}

	// Static Pages
	{
		// js, css, etc.
		app.Group("/static").Group(cache.Rev(), cache.Set(cache.ForRevvedResource)).StaticFS("/", http.FS(assets.Static()))

		app.GET("/", html("index.html"))
		app.GET("/users/register", html("register.html"))
		app.GET("/users/login", html("login.html"))
		app.GET("/posts", html("posts.html"))
		app.GET("/admin", html("admin.html"))
		app.GET("/admin/iam", redirect("/admin/iam/users"))
		app.GET("/admin/iam/users", html("admin/iam/users.html"))
		app.GET("/admin/iam/groups", html("admin/iam/groups.html"))
		app.GET("/admin/iam/roles", html("admin/iam/roles.html"))

		// bootstrap
		app.GET("/bootstarp", bootstrap.IssueBootstrapToken, html("bootstrap.html"))

		// or for SPA(single page application), client side routing
		// app.Use(AbortIfHasPrefix("/api"), server.static("/index.html"))

		// for dev
		// app.GET("/dev/palette", html("dev/palette.html"))
	}

	noRoute(func(c *gin.Context) {
		for accept := range strings.SplitSeq(c.Request.Header.Get("Accept"), ",") {
			t, _, e := mime.ParseMediaType(accept)
			if e != nil {
				continue
			}

			switch t {
			case "application/json":
				c.Status(http.StatusNotFound)
				return
			case "text/html", "*/*":
				c.HTML(http.StatusNotFound, "errors/not-found.html", c)
				return
			case "text/plain":
				c.String(http.StatusNotFound, "not found")
				return
			}
		}
	})
}
func api(fn func(c *gin.Context) error) gin.HandlerFunc {

	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			c.Error(err)
			c.Abort()
		}
	}
}

type ServerSentEventErrorData struct {
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Status     int            `json:"status,omitempty"`
	Detail     string         `json:"detail,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Extensions map[string]any `json:"-"`
}

func sse(fn func(c *gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := http.NewResponseController(c.Writer)
		rc.SetWriteDeadline(time.Time{})
		rc.SetReadDeadline(time.Time{})

		if err := fn(c); err != nil {
			c.SSEvent("error", ServerSentEventErrorData{
				Type:   "about:blank",
				Title:  "Error",
				Detail: err.Error(),
			})
		}
	}
}

func bodyReaderTweak(c *gin.Context) {
	// body 를 여러번 읽지 못하는 것을 대응 하기 위한 tweak
	body := c.Request.Body

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, body); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	// should i use `gin.BodyBytesKey`? https://stackoverflow.com/questions/62736851/go-gin-read-request-body-many-times

	// Data 는 buf에서 읽고, close 는 원래것에서 호출
	c.Request.Body = struct {
		io.Reader
		io.Closer
	}{
		Reader: bytes.NewReader(buf.Bytes()),
		Closer: body,
	}

}
