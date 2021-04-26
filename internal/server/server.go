package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/auth"
)

type Config struct {
	Bind     string
	GRPCBind string
	Key      string
	DBPath   string
	Salt     string
	InitUser map[string]string
}

func NewConfig() Config {
	return Config{
		InitUser: map[string]string{},
	}
}

type Server struct {
	conf *Config
	db   *gorm.DB
	auth *auth.Manager
}

func Run(conf *Config) error {
	server := &Server{
		conf: conf,
	}

	// init components
	if err := server.initDB(); err != nil {
		return errors.Wrap(err, "init server failed")
	}
	if err := server.initAuth(); err != nil {
		return errors.Wrap(err, "init server failed")
	}

	// run servers
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(server.RunHTTPServer)
	eg.Go(server.RunGRPCServer)

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
