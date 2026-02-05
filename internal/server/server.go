package server

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/hjson/hjson-go/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/bluemir/0xC0DE/assets"
	"github.com/bluemir/0xC0DE/internal/server/backend"
	"github.com/bluemir/0xC0DE/internal/server/store"
)

type Args struct {
	ServiceHttpBind string
	Cert            CertConfig
	GRPCBind        string
	AdminHttpBind   string
	ConfigFilePath  string

	DBPath string
	Salt   string
}

type Config struct {
	Backend backend.Config
}

type Server struct {
	backends *backend.Backends
}

func Run(ctx context.Context, args *Args) error {
	if err := assets.CheckDevAssets(); err != nil {
		return err
	}

	conf, err := readCofigFile(args.ConfigFilePath)
	if err != nil {
		return errors.Wrapf(err, "config file not exist. path: %s", args.ConfigFilePath)
	}

	// pass cmd to config
	conf.Backend.Auth.Salt = args.Salt

	db, err := store.Initialize(ctx, args.DBPath)
	if err != nil {
		return err
	}

	bs, err := backend.Initialize(ctx, &conf.Backend, db)
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

	gwHandler, err := server.grpcGatewayHandler(ctx, args.GRPCBind)
	if err != nil {
		return err
	}

	certs, err := args.Cert.Load()
	if err != nil {
		return err
	}

	tlsConfig, err := getTLSConfig(certs, nil)
	if err != nil {
		return err
	}

	// run servers
	eg, nCtx := errgroup.WithContext(ctx)
	eg.Go(server.RunServiceHTTPServer(nCtx, args.ServiceHttpBind, tlsConfig, gwHandler))
	eg.Go(server.RunAdminHTTPServer(nCtx, args.AdminHttpBind))
	eg.Go(server.RunGRPCServer(nCtx, args.GRPCBind))

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

func readCofigFile(configFilePath string) (*Config, error) {
	conf := Config{}

	buf, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch filepath.Ext(configFilePath) {
	case ".yaml", ".yml":
		logrus.Info("parse as yaml")
		if err := yaml.Unmarshal(buf, &conf); err != nil {
			return nil, errors.WithStack(err)
		}
	case ".json", ".hjson":
		logrus.Info("parse as hjson")
		if err := hjson.Unmarshal(buf, &conf); err != nil {
			return nil, errors.WithStack(err)
		}
	default:
		return nil, errors.Errorf("unknown ext: %s", filepath.Ext(configFilePath))
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		buf, _ := hjson.Marshal(conf)
		logrus.Debugf("\n%s", string(buf))
	}

	return &conf, nil
}
