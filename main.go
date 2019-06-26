package main

import (
	"os"

	"github.com/codingconcepts/env"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/0xC0DE/pkg/server"
)

var VERSION string

func main() {
	// log
	if level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL")); err != nil {
		logrus.Warn("unknown log level. using default level(info)")
	} else {
		logrus.SetLevel(level)
	}

	conf := &server.Config{}

	if err := env.Set(conf); err != nil {
		logrus.Fatal(err)
		return
	}

	cli := kingpin.New("0xC0DE", "main code")

	cli.Flag("debug", "enable debug mode").BoolVar(&conf.Debug)
	cli.Flag("bind", "bind address").StringVar(&conf.Bind)

	cli.Version(VERSION)

	kingpin.MustParse(cli.Parse(os.Args[1:]))

	if err := server.Run(conf); err != nil {
		logrus.Fatal(err)
	}
}
