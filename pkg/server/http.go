package server

import (
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
		WithField("from", "gin").
		WriterLevel(logrus.DebugLevel)
	defer writer.Close()

	app.Use(gin.LoggerWithWriter(writer))
	app.Use(gin.Recovery())

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
