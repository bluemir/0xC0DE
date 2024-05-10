package server

import (
	"context"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/server/graceful"
	"github.com/bluemir/0xC0DE/internal/server/handler"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth"
	"github.com/bluemir/0xC0DE/internal/server/middleware/errs"
	"github.com/bluemir/0xC0DE/internal/server/middleware/prom"
)

func (server *Server) RunServiceHTTPServer(ctx context.Context, bind string, certs *CertConfig, extra ...gin.HandlerFunc) func() error {

	return func() error {
		// starting http server
		app := gin.New()

		// add template
		if html, err := NewRenderer(); err != nil {
			return err
		} else {
			app.SetHTMLTemplate(html)
		}

		// setup Logger
		writer := logrus.
			WithFields(logrus.Fields{}).
			WriterLevel(logrus.DebugLevel)
		defer writer.Close()
		app.Use(gin.LoggerWithWriter(writer))

		// sessions
		store := cookie.NewStore([]byte(server.salt))
		app.Use(sessions.Sessions("0xC0DE_session", store))

		app.Use(gin.Recovery())

		app.Use(location.Default(), fixURL)

		app.Use(errs.Middleware)

		app.Use(auth.Middleware(server.backends.Auth))

		app.Use(handler.Inject(server.backends))

		// prometheus for monitoring
		app.Use(prom.Metrics())

		// handle routes
		server.routes(app, app.NoRoute)

		// GRPC Gateway
		app.Use(extra...)
		// app.Group("/grpc/*any", extra...)

		return graceful.Run(ctx, bind, app, graceful.WithCert(certs))
	}
}
