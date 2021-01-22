package cmd

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/0xC0DE/pkg/client"
	"github.com/bluemir/0xC0DE/pkg/server"
	"github.com/bluemir/0xC0DE/pkg/util"
)

const (
	describe        = ``
	defaultLogLevel = logrus.InfoLevel
)

func Run(AppName string, Version string) error {
	conf := struct {
		server   server.Config
		client   client.Config
		logLevel int
	}{}

	app := kingpin.New(AppName, describe)
	app.Version(Version)

	app.Flag("verbose", "Log level").
		Short('v').
		CounterVar(&conf.logLevel)

	serverCmd := app.Command("server", "server")
	{
		serverCmd.Flag("bind", "bind").
			Default(":8080").
			StringVar(&conf.server.Bind)
		serverCmd.Flag("key", "key(default: random string)").
			Default(util.RandomString(16)).PlaceHolder("KEY").
			Envar(strings.ToUpper(AppName) + "_KEY").
			StringVar(&conf.server.Key)
		serverCmd.Flag("db-path", "db path").
			Default(":memory:").
			StringVar(&conf.server.DBPath)
		serverCmd.Flag("grpc-bind", "grpc bind").
			Default(":3277").
			StringVar(&conf.server.GRPCBind)
	}

	clientCmd := app.Command("client", "client")
	{
		clientCmd.Flag("endpoint", "endpoint").
			Default("localhost:3277").
			StringVar(&conf.client.Endpoint)
	}

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	level := logrus.Level(conf.logLevel) + defaultLogLevel
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.Infof("logrus level: %s", level)

	switch cmd {
	case serverCmd.FullCommand():
		return server.Run(&conf.server)
	case clientCmd.FullCommand():
		return client.Run(&conf.client)
	default:
		return errors.New("not implements command")
	}
}