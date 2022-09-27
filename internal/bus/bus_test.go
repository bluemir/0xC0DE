package bus

import (
	"context"
	"testing"
	"time"
)

type EventContext struct {
	*testing.T
}

func (ctx EventContext) Print(msg string, args ...interface{}) {
	ctx.Logf(msg, args...)
}

type ClickEvent struct {
}
type KeyDownEvent struct {
}

type Any interface{}
type DumpHandler struct{}

func (DumpHandler) Handle(evt Event[EventContext]) {
	switch e := evt.Detail.(type) {
	case ClickEvent:
		evt.Context.Print("ClickEvent: %#v", e)
	default:
		evt.Context.Print("others: %#v", e)
	}
}

func TestBus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	bus, err := NewBus[EventContext](ctx)
	if err != nil {
		t.Error(err)
		return
	}
	bus.AddEventListener("*", DumpHandler{})

	bus.FireEvent(Event[EventContext]{
		Kind:    "click",
		Context: EventContext{t},
		Detail:  ClickEvent{},
	})

	bus.FireEvent(Event[EventContext]{
		Kind:    "click",
		Context: EventContext{t},
		Detail:  KeyDownEvent{},
	})
}
