package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
