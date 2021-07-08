package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/cmd"
	"github.com/bluemir/0xC0DE/internal/buildinfo"
)

var Version string
var AppName string
var Time string

func main() {
	buildinfo.AppName = AppName
	buildinfo.Version = Version
	buildinfo.Time = Time

	if err := cmd.Run(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
