package client

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/bluemir/0xC0DE/internal/client"
)

func Register(cmd *kingpin.CmdClause) {
	conf := client.Config{}
	cmd.Flag("endpoint", "endpoint").
		Default("localhost:3277").
		StringVar(&conf.Endpoint)
	cmd.Action(func(c *kingpin.ParseContext) error {
		return client.Run(&conf)
	})
}
