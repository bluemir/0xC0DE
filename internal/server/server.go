package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/bluemir/0xC0DE/internal/server/handler"
)

type Config struct {
	Bind     string
	KeyFile  string
	CertFile string
	GRPCBind string
	DBPath   string
	Salt     string
	Seed     string
	InitUser map[string]string
}

func NewConfig() Config {
	return Config{
		InitUser: map[string]string{},
	}
}

type Server struct {
	conf    *Config
	db      *gorm.DB
	auth    *auth.Manager
	handler *handler.Handler
	etag    string
}

func Run(ctx context.Context, conf *Config) error {
	server := &Server{
		conf: conf,
	}

	// init components
	if err := server.initEtag(); err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	if err := server.initDB(); err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	if err := server.initAuth(); err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	if h, err := handler.New(server.db); err != nil {
		return errors.Wrapf(err, "init handler failed")
	} else {
		server.handler = h
	}

	// run servers
	eg, nCtx := errgroup.WithContext(ctx)
	eg.Go(server.RunHTTPServer(nCtx))
	eg.Go(server.RunGRPCServer(nCtx))

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
