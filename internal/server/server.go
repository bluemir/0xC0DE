package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/bluemir/0xC0DE/internal/server/handler"
)

type Config struct {
	Bind     string
	Cert     CertConfig
	GRPCBind string
	DBPath   string
	Salt     string
	Seed     string
	InitUser map[string]string
}
type CertConfig struct {
	CertFile string
	KeyFile  string
}

func NewConfig() Config {
	return Config{
		InitUser: map[string]string{},
	}
}

type Server struct {
	salt    string
	auth    *auth.Manager
	handler *handler.Handler
}

func Run(ctx context.Context, conf *Config) error {

	// init components
	db, err := initDB(conf.DBPath)
	if err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	authManager, err := initAuth(db, conf.Salt, conf.InitUser)
	if err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	h, err := handler.New(db)
	if err != nil {
		return err
	}

	// option 1. single handler, multiple backend
	// route -> handler -> backend -> db
	//                  -> backend -> db
	// option 2. multiple handler or direct backend
	// route -> handler -> db
	//       -> handler -> backend -> db
	//       -> backend -> db
	server := &Server{
		salt: conf.Salt,
		auth: authManager,

		handler: h,
	}

	gwHandler, err := server.grpcGatewayHandler(ctx, conf.GRPCBind)
	if err != nil {
		return err
	}

	// run servers
	eg, nCtx := errgroup.WithContext(ctx)
	eg.Go(server.RunHTTPServer(nCtx, conf.Bind, conf.GetCertConfig(), gwHandler))
	eg.Go(server.RunGRPCServer(nCtx, conf.GRPCBind))

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
func (conf *Config) GetCertConfig() *CertConfig {
	if conf.Cert.CertFile == "" && conf.Cert.KeyFile == "" {
		return nil
	}
	return &conf.Cert
}
