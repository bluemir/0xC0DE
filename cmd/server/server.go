package server

import (
	"context"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
	"github.com/bluemir/0xC0DE/internal/server"
	"github.com/bluemir/0xC0DE/internal/util"
)

func Register(cmd *kingpin.CmdClause) {
	conf := server.NewConfig()

	cmd.Flag("bind", "bind").
		Default(":8080").
		StringVar(&conf.ServiceHttpBind)
	cmd.Flag("cert", "cert file").
		StringVar(&conf.Cert.CertFile)
	cmd.Flag("key", "key file").
		StringVar(&conf.Cert.KeyFile)
	cmd.Flag("admin-bind", "bind for admin endpoint(metric, swagger, pprof)").
		Default(":8000").
		StringVar(&conf.AdminHttpBind)

	cmd.Flag("db-path", "db path").
		Default(":memory:").
		StringVar(&conf.DBPath)
	cmd.Flag("salt", "salt(default: random string)").
		Default(util.RandomString(16)).PlaceHolder("KEY").
		Envar(strings.ToUpper(buildinfo.AppName) + "_SALT").
		StringVar(&conf.Salt)
	cmd.Flag("grpc-bind", "grpc bind").
		Default(":3277").
		StringVar(&conf.GRPCBind)
	cmd.Flag("init-user", "initial user").
		StringMapVar(&conf.InitUser)
	cmd.Flag("config", "config file").
		Short('c').StringVar(&conf.ConfigFilePath)
	cmd.Action(func(*kingpin.ParseContext) error {
		logrus.Trace("called")

		ctx, stop := signal.NotifyContext(context.Background(),
			syscall.SIGTERM,
			syscall.SIGINT,
		)
		defer stop()

		return server.Run(ctx, &conf)
	})
}
