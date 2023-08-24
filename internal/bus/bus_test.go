package bus

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	bus.AddAllEventListener(h)

	bus.FireEvent("click", EventContext{t})
	bus.FireEvent("keydown", EventContext{t})
}
func TestBusCall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bus, err := NewBus(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	ch := bus.WatchAllEvent(ctx.Done())

	counter := runEventHandler(ch)

	bus.FireEvent("other", nil)
	bus.FireEvent("other", nil)
	bus.FireEvent("other", nil)

	time.Sleep(1 * time.Second)

	assert.Equal(t, 3, *counter)

}
func runEventHandler(ch <-chan Event) *int {
	c := 0
	go func() {
		for evt := range ch {
			c += 1

			logrus.Trace(evt)
		}
	}()
	return &c
}
