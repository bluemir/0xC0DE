package events_test

import (
	"github.com/bluemir/0xC0DE/internal/events"
	"github.com/sirupsen/logrus"
)

type FowardHandler struct {
	to string
}

func (h FowardHandler) Handle(ctx events.Context, evt events.Event) {
	logrus.Trace(evt)
	ctx.FireEvent(h.to, evt.Detail)
}
