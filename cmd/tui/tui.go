package tui

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
	"github.com/bluemir/0xC0DE/internal/tui"
)

func Register(cmd *kingpin.CmdClause) {

	cmd.Action(func(*kingpin.ParseContext) error {
		logrus.Trace("called")

		logrus.Infof("Build mode: %s", buildinfo.BuildMode)

		ctx, stop := signal.NotifyContext(context.Background(),
			syscall.SIGTERM,
			syscall.SIGINT,
		)
		defer stop()

		return tui.Run(ctx)
	})
}
