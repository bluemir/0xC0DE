package pubsub_test

import (
	"github.com/bluemir/0xC0DE/internal/pubsub/v1"
	"github.com/sirupsen/logrus"
)

type FowardHandler struct {
	to  string
	Hub pubsub.IHub
}

func (h FowardHandler) Handle(evt pubsub.Message) error {
	logrus.Trace(evt)
	h.Hub.Publish(h.to, evt.Detail)
	return nil
}
