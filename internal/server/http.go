package server

import (
	"encoding/gob"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/gin-contrib/location"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (server *Server) RunHTTPServer() error {
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

	return app.Run(server.conf.Bind)
}

func fixURL(c *gin.Context) {
	url := location.Get(c)

	// QUESTION is it right?
	c.Request.URL.Scheme = url.Scheme
	c.Request.URL.Host = url.Host
}
