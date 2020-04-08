package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/0xC0DE/pkg/server"
	"github.com/bluemir/0xC0DE/pkg/util"
)

var Version string
var AppName string

func main() {
	logLevel := 0
	conf := struct {
		server server.Config
	}{}

	app := kingpin.New(AppName, AppName+" describe")
	app.Version(Version)

	app.Flag("verbose", "Log level").Short('v').CounterVar(&logLevel)

	serverCmd := app.Command("server", "server")
	serverCmd.Flag("bind", "bind").
		Default(":8080").
		StringVar(&conf.server.Bind)
	serverCmd.Flag("key", "key(default: random string)").
		Default(util.RandomString(16)).PlaceHolder("KEY").
		Envar(strings.ToUpper(AppName) + "_KEY").
		StringVar(&conf.server.Key)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	level := logrus.Level(logLevel) + logrus.ErrorLevel // default is error level
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.Infof("error level: %s", level)

	Run := func(cmd string) error {
		switch cmd {

		case serverCmd.FullCommand():
			return server.Run(&conf.server)
		}
		return nil
	}

	if err := Run(cmd); err != nil {
		logrus.Error(err)
	}
}
