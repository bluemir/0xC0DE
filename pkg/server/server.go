package server

import (
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/pkg/dist"
)

type Config struct {
	Debug bool
	Bind  string
}

func Run(conf *Config) error {
	logrus.Debugf("%#v", conf)

	logrus.Tracef("%#v", dist.Files)

	return nil
}
