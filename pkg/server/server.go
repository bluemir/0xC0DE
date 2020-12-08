package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Config struct {
	Bind   string
	Key    string
	DBPath string
}
type Server struct {
	conf *Config
	grpc *grpc.Server
	db   *gorm.DB
}

func Run(conf *Config) error {
	server := &Server{
		conf: conf,
	}

	server.initGrpcService()
	server.initDB()

	app := gin.New()

	// add template
	if html, err := NewRenderer(); err != nil {
		return err
	} else {
		app.SetHTMLTemplate(html)
	}

	// setup Logger
	writer := logrus.New().Writer()
	defer writer.Close()

	app.Use(gin.LoggerWithWriter(writer))
	app.Use(gin.Recovery())

	// handle static

	server.routes(app)

	return app.Run(conf.Bind)
}
