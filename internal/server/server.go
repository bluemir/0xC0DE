package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	backends "github.com/bluemir/0xC0DE/internal/server/backend"
)

type Config struct {
	HttpBind  string
	Cert      CertConfig
	GRPCBind  string
	PprofBind string

	backends.Args
}
type CertConfig struct {
	CertFile string
	KeyFile  string
}

func NewConfig() Config {
	return Config{
		Args: backends.NewArgs(),
	}
}

type Server struct {
	salt string

	backends *backends.Backends
}

func Run(ctx context.Context, conf *Config) error {
	bs, err := backends.Initialize(ctx, &conf.Args)
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
	eg.Go(server.RunHTTPServer(nCtx, conf.HttpBind, conf.GetCertConfig(), gwHandler))
	//eg.Go(server.RunHTTPServer(nCtx, conf.Bind, conf.GetCertConfig()))
	eg.Go(server.RunGRPCServer(nCtx, conf.GRPCBind))
	eg.Go(server.RunPprofServer(nCtx, conf.PprofBind))

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
