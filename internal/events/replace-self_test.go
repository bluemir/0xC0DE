package events_test

import "github.com/bluemir/0xC0DE/internal/events"

type ReplaceSelfHandler struct{}

func (ReplaceSelfHandler) Handle(ctx events.Context, evt events.Event) {
	ctx.RemoveEventHandler("do", ReplaceSelfHandler{})
	ctx.AddEventHandler("do", FowardHandler{to: "done"})
}
