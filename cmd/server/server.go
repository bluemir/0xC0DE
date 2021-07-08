package server

import (
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
	"github.com/bluemir/0xC0DE/internal/server"
	"github.com/bluemir/0xC0DE/internal/util"
)

func Register(cmd *kingpin.CmdClause) {
	conf := server.NewConfig()

	cmd.Flag("bind", "bind").
		Default(":8080").
		StringVar(&conf.Bind)
	cmd.Flag("key", "key(default: random string)").
		Default(util.RandomString(16)).PlaceHolder("KEY").
		Envar(strings.ToUpper(buildinfo.AppName) + "_KEY").
		StringVar(&conf.Key)
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
	cmd.Action(func(*kingpin.ParseContext) error {
		logrus.Trace("called")
		return server.Run(&conf)
	})
}
