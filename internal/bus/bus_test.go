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

func DumpHandler() chan<- Event[EventContext] {
	ch := make(chan Event[EventContext])
	go func() {
		for evt := range ch {
			switch evt.Kind {
			case "click":
				evt.Detail.Print("ClickEvent: %#v", evt)
			default:
				evt.Detail.Print("others: %#v", evt)
			}
		}
	}()
	return ch
}

func TestBus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	bus, err := NewBus[EventContext](ctx)
	if err != nil {
		t.Error(err)
		return
	}
	h := DumpHandler()
	bus.AddEventListener("*", h)

	bus.FireEvent(Event[EventContext]{
		Kind:   "click",
		Detail: EventContext{t},
	})

	bus.FireEvent(Event[EventContext]{
		Kind:   "keydown",
		Detail: EventContext{t},
	})
}
