package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	clientCmd "github.com/bluemir/0xC0DE/cmd/client"
	serverCmd "github.com/bluemir/0xC0DE/cmd/server"
)

const (
	describe        = ``
	defaultLogLevel = logrus.WarnLevel
)

func Run(AppName string, Version string) error {
	conf := struct {
		logLevel  int
		logFormat string
	}{}

	app := kingpin.New(AppName, describe)
	app.Version(Version)

	app.Flag("verbose", "Log level").
		Short('v').
		CounterVar(&conf.logLevel)
	app.Flag("log-format", "Log format").
		StringVar(&conf.logFormat)
	app.PreAction(func(*kingpin.ParseContext) error {
		level := logrus.Level(conf.logLevel) + defaultLogLevel
		logrus.SetOutput(os.Stderr)
		logrus.SetLevel(level)
		logrus.SetReportCaller(true)
		logrus.Infof("logrus level: %s", level)

		switch conf.logFormat {
		case "text-color":
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		case "text":
			logrus.SetFormatter(&logrus.TextFormatter{})
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "":
			// do nothing. it means smart.
		default:
			return errors.Errorf("unknown log format")
		}

		return nil
	})

	serverCmd.Register(app.Command("server", "server"), AppName)

	clientCmd.Register(app.Command("client", "client"), AppName)

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	logrus.Debug(cmd)
	return nil
}
