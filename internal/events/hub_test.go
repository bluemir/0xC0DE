package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/events"
)

type key struct{}

var tkey key = struct{}{}

func testContext(t *testing.T, timeout time.Duration) (context.Context, func()) {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)

	ctx := context.Background()
	ctx = context.WithValue(ctx, tkey, t)
	return context.WithTimeout(ctx, timeout)
}
func From(ctx context.Context) *testing.T {
	return ctx.Value(tkey).(*testing.T)
}

func TestSendEvent(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := events.NewHub(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	h := &CounterHandler{}

	hub.AddAllEventHandler(h)

	hub.FireEvent("test", nil)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, h.GetCount())
}
func TestSendMultiple(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := events.NewHub(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	recoder := NewRecoder(ctx, hub)

	hub.FireEvent("test-1", nil)
	hub.FireEvent("test-2", nil)
	hub.FireEvent("test-3", nil)
	hub.FireEvent("test-4", nil)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 4, len(recoder.recodes))
	assert.Equal(t, []string{
		"test-1",
		"test-2",
		"test-3",
		"test-4",
	}, recoder.History())
}

func TestAddEventHandlerWithNull(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := events.NewHub(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	hub.AddEventHandler("test", events.Handler(nil))

	hub.FireEvent("test", nil)
}

func TestEventKind(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := events.NewHub(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	counter := &CounterHandler{}
	hub.AddEventHandler("test-1", counter)

	hub.FireEvent("test-2", nil)

	assert.Equal(t, 0, counter.GetCount())
}

func TestFireEventInsideEventHandler(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := events.NewHub(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	recoder := NewRecoder(ctx, hub)

	hub.AddEventHandler("button.down", FowardHandler{"click"})

	hub.FireEvent("button.down", nil)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, []string{
		"button.down",
		"click",
	}, recoder.History())
}
