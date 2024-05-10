package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/bluemir/0xC0DE/internal/server/backend"
	"github.com/bluemir/0xC0DE/internal/server/graceful"
)

type Config struct {
	ServiceHttpBind string
	Cert            CertConfig
	GRPCBind        string
	AdminHttpBind   string

	backend.Args
}
type CertConfig = graceful.CertConfig

func NewConfig() Config {
	return Config{
		Args: backend.NewArgs(),
	}
}

type Server struct {
	salt string

	backends *backend.Backends
}

func Run(ctx context.Context, conf *Config) error {
	bs, err := backend.Initialize(ctx, &conf.Args)
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

		// backends
		backends: bs,
	}

	gwHandler, err := server.grpcGatewayHandler(ctx, conf.GRPCBind)
	if err != nil {
		return err
	}

	// run servers
	eg, nCtx := errgroup.WithContext(ctx)
	eg.Go(server.RunServiceHTTPServer(nCtx, conf.ServiceHttpBind, conf.GetCertConfig(), gwHandler))
	eg.Go(server.RunAdminHTTPServer(nCtx, conf.AdminHttpBind))
	//eg.Go(server.RunHTTPServer(nCtx, conf.Bind, conf.GetCertConfig()))
	eg.Go(server.RunGRPCServer(nCtx, conf.GRPCBind))

	// TODO run grpc, http, https, http2https redirect servers by config

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
