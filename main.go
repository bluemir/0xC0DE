package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/cmd"
)

var Version string
var AppName string

func main() {
	if err := cmd.Run(AppName, Version); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
