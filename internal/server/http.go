package server

import (
	"context"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/gin-contrib/location"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (server *Server) RunHTTPServer(ctx context.Context) func() error {
	return func() error {
		// starting http server
		app := gin.New()

		// add template
		if html, err := NewRenderer(); err != nil {
			return errors.WithStack(err)
		} else {
			app.SetHTMLTemplate(html)
		}

		// setup Logger
		writer := logrus.
			WithFields(logrus.Fields{}).
			WriterLevel(logrus.DebugLevel)
		defer writer.Close()

		// sessions
		gob.Register(&auth.Token{})
		store := cookie.NewStore([]byte(server.conf.Salt))
		app.Use(sessions.Sessions("0xC0DE_session", store))

		app.Use(gin.LoggerWithWriter(writer))
		app.Use(gin.Recovery())

		app.Use(location.Default())
		app.Use(fixURL)

		// handle routes
		server.routes(app)

		// GRPC Gateway
		mw, err := server.grpcGatewayMiddleware()
		if err != nil {
			return errors.WithStack(err)
		}
		app.Use(mw)

		// setup graceful server
		// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-with-context/server.go
		s := http.Server{
			Addr:    server.conf.Bind,
			Handler: app,
		}

		errc := make(chan error)
		go func() {
			defer close(errc)

			logrus.Infof("Listening and serving HTTP on '%s'", server.conf.Bind)
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errc <- err
			}
		}()

		select {
		case <-ctx.Done():
			logrus.Warn("shutting down gracefully, press Ctrl+C again to force")
		case err := <-errc:
			logrus.Errorf("listen: %s\n", err)
		}

		// nCtx for shutdown timeout only
		nCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(nCtx); err != nil {
			return errors.Wrapf(err, "Server forced to shutdown: ")
		}

		return nil
	}

}

func fixURL(c *gin.Context) {
	url := location.Get(c)

	// QUESTION is it right?
	c.Request.URL.Scheme = url.Scheme
	c.Request.URL.Host = url.Host
}
