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

func DumpHandler() chan<- Event {
	ch := make(chan Event)
	go func() {
		for evt := range ch {
			switch evt.Kind {
			case "click":
				evt.Detail.(EventContext).Print("ClickEvent: %#v", evt)
			default:
				evt.Detail.(EventContext).Print("others: %#v", evt)
			}
		}
	}()
	return ch
}

func TestBus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	bus, err := NewBus(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	h := DumpHandler()
	bus.AddEventListener("*", h)

	bus.FireEvent("click", EventContext{t})
	bus.FireEvent("keydown", EventContext{t})
}
