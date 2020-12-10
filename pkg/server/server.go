package server

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/pkg/util"
)

type Config struct {
	Bind     string
	GRPCBind string
	Key      string
	DBPath   string
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

	// init components
	if err := util.MergeErrors(
		server.initGrpcService(),
		server.initDB(),
	); err != nil {
		return errors.Wrap(err, "init server failed")
	}

	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		// starting http server
		app := gin.New()

		// add template
		if html, err := NewRenderer(); err != nil {
			return errors.WithStack(err)
		} else {
			app.SetHTMLTemplate(html)
		}

		// setup Logger
		writer := logrus.New().Writer()
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

		return app.Run(conf.Bind)
	})
	eg.Go(func() error {
		// Starting grpc Server
		lis, err := net.Listen("tcp", conf.GRPCBind)
		if err != nil {
			logrus.Fatalf("failed to listen: %v", err)
		}

		return server.grpc.Serve(lis)
	})

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
