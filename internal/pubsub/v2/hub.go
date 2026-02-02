package pubsub

import (
	"context"
	"reflect"
	"time"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type IHub interface {
	Publish(ctx context.Context, detail any)
	AddHandler(kind any, handler Handler)
	RemoveHandler(kind any, handler Handler)
	Watch(kind any, done <-chan struct{}, opts ...WatchOption) <-chan Event
	WatchAll(done <-chan struct{}, opts ...WatchOption) <-chan Event
}

type watchConfig struct {
	bufferSize int
}

type WatchOption func(*watchConfig)

func WithBuffer(n int) WatchOption {
	return func(c *watchConfig) {
		c.bufferSize = n
	}
}

var _ IHub = (*Hub)(nil)

type Hub struct {
	handlers    datastruct.Map[reflect.Type, datastruct.Set[Handler]]
	broadcaster *Broadcaster[Event]
}

func NewHub(ctx context.Context) (*Hub, error) {
	return &Hub{
		broadcaster: NewBroadcaster[Event](),
	}, nil
}

type keyTypeHub struct{}

var keyHub = keyTypeHub{}

func HubFrom(ctx context.Context) *Hub {
	return ctx.Value(keyHub).(*Hub)
}

func (h chanEventHandler) Handle(ctx context.Context, evt Event) error {
	h.ch <- evt
	return nil
}

type chanEventHandler struct {
	ch chan<- Event
}

func (hub *Hub) Publish(ctx context.Context, detail any) {
	kind := reflect.TypeOf(detail)

	ctx = context.WithValue(ctx, keyHub, hub)
	evt := Event{
		Context: ctx,
		Id:      xid.New().String(),
		At:      time.Now(),
		Detail:  detail,
		Kind:    kind.String(),
	}

	hub.broadcaster.Broadcast(evt)

	handlers, ok := hub.handlers.Get(kind)
	if !ok {
		return
	}

	snapshot := []Handler{}
	if err := handlers.Range(func(handler Handler) error {
		snapshot = append(snapshot, handler)
		return nil
	}); err != nil {
		logrus.Debug(err)
	}

	for _, handler := range snapshot {
		if err := handler.Handle(ctx, evt); err != nil {
			logrus.Debug(err)
		}
	}

}
func (hub *Hub) AddHandler(kind any, handler Handler) {
	handlers, _ := hub.handlers.GetOrSet(reflect.TypeOf(kind), datastruct.NewSet[Handler]())
	handlers.Add(handler)
}
func (hub *Hub) RemoveHandler(kind any, handler Handler) {
	handlers, _ := hub.handlers.GetOrSet(reflect.TypeOf(kind), datastruct.NewSet[Handler]())
	handlers.Remove(handler)
}
func (hub *Hub) Watch(kind any, done <-chan struct{}, opts ...WatchOption) <-chan Event {
	conf := watchConfig{
		bufferSize: -1,
	}
	for _, opt := range opts {
		opt(&conf)
	}

	var ch chan Event
	if conf.bufferSize >= 0 {
		ch = make(chan Event, conf.bufferSize)
	} else {
		ch = make(chan Event)
	}

	h := chanEventHandler{
		ch: ch,
	}

	hub.AddHandler(kind, h)
	go func() {
		<-done
		hub.RemoveHandler(kind, h)
		close(ch)
	}()

	if conf.bufferSize >= 0 {
		return ch
	}

	// Use DynamicChan to prevent deadlock during recursive publishing
	return datastruct.DynamicChan(ch)
}
func (hub *Hub) WatchAll(done <-chan struct{}, opts ...WatchOption) <-chan Event {
	return hub.broadcaster.Watch(done, opts...)
}
