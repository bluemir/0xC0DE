package server

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/bluemir/0xC0DE/internal/server/backend"
)

type Config struct {
	ServiceHttpBind string
	Cert            CertConfig
	GRPCBind        string
	AdminHttpBind   string

	backend.Args
}

func NewConfig() Config {
	return Config{
		Args: backend.NewArgs(),
	}
}

type Server struct {
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
		// backends
		backends: bs,
	}

	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		gin.SetMode(gin.ReleaseMode)
	}

	gwHandler, err := server.grpcGatewayHandler(ctx, conf.GRPCBind)
	if err != nil {
		return err
	}

	certs, err := conf.Cert.Load()
	if err != nil {
		return err
	}

	tlsConfig, err := getTLSConfig(certs, nil)
	if err != nil {
		return err
	}

	// run servers
	eg, nCtx := errgroup.WithContext(ctx)
	eg.Go(server.RunServiceHTTPServer(nCtx, conf.ServiceHttpBind, tlsConfig, gwHandler))
	eg.Go(server.RunAdminHTTPServer(nCtx, conf.AdminHttpBind))
	eg.Go(server.RunGRPCServer(nCtx, conf.GRPCBind))

	// TODO run grpc, http, https, http2https redirect servers by config

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type CertConfig struct {
	CertFile string
	KeyFile  string
}

func (cert *CertConfig) Load() (*tls.Certificate, error) {
	if cert == nil {
		return nil, nil
	}
	if cert.CertFile == "" || cert.KeyFile == "" {
		return nil, nil
	}
	c, err := tls.LoadX509KeyPair(cert.CertFile, cert.KeyFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &c, nil
}
func getTLSConfig(serverCert *tls.Certificate, clientAuthCACert *tls.Certificate) (*tls.Config, error) {
	if serverCert == nil {
		return nil, nil
	}
	conf := tls.Config{
		Certificates: []tls.Certificate{*serverCert},
	}

	if clientAuthCACert != nil {
		conf.ClientAuth = tls.VerifyClientCertIfGiven
		conf.ClientCAs = x509.NewCertPool()

		for _, der := range clientAuthCACert.Certificate {
			c, err := x509.ParseCertificate(der)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			conf.ClientCAs.AddCert(c)
		}
	}

	return &conf, nil
}
