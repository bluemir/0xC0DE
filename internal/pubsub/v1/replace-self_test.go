package pubsub_test

import "github.com/bluemir/0xC0DE/internal/pubsub/v1"

type ReplaceSelfHandler struct {
	Hub pubsub.IHub
}

func (h *ReplaceSelfHandler) Handle(evt pubsub.Message) {
	h.Hub.RemoveHandler("do", h)
	h.Hub.AddHandler("do", FowardHandler{to: "done", Hub: h.Hub})
}
