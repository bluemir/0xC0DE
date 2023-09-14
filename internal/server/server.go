package server

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/bluemir/0xC0DE/internal/bus"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/posts"
)

type Config struct {
	HttpBind string
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
	salt string

	auth  *auth.Manager
	bus   *bus.Bus
	posts *posts.Manager
}

func Run(ctx context.Context, conf *Config) error {
	events, err := bus.NewBus(ctx)
	if err != nil {
		return err
	}
	// init components
	db, err := initDB(conf.DBPath)
	if err != nil {
		return errors.Wrapf(err, "init server failed")
	}
	authManager, err := initAuth(db, conf.Salt, conf.InitUser)
	if err != nil {
		return errors.Wrapf(err, "init server failed")
	}

	postManager, err := posts.New(db, events)
	if err != nil {
		return errors.Wrapf(err, "init post manager failed")
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
		auth:  authManager,
		bus:   events,
		posts: postManager,
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
