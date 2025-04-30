package pubsub

import (
	"context"
	"reflect"
	"time"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/rs/xid"
)

type IHub interface {
	Publish(ctx context.Context, detail any)
	AddHandler(kind any, handler Handler)
	RemoveHandler(kind any, handler Handler)
	Watch(kind any, done <-chan struct{}) <-chan Event
	WatchAll(done <-chan struct{}) <-chan Event
}

var _ IHub = (*Hub)(nil)

type Hub struct {
	handlers datastruct.Map[reflect.Type, datastruct.Set[Handler]]

	all datastruct.Set[chan<- Event]
}

func NewHub(ctx context.Context) (*Hub, error) {
	return &Hub{}, nil
}

func (hub *Hub) Publish(ctx context.Context, detail any) {
	kind := reflect.TypeOf(detail)

	handlers, ok := hub.handlers.Get(kind)
	if !ok {
		return
	}
	evt := Event{
		Context: ctx,
		Id:      xid.New().String(),
		At:      time.Now(),
		Detail:  detail,
		Kind:    kind.String(),
	}

	for _, handler := range handlers.List() {
		handler.Handle(ctx, evt)
	}

	for _, ch := range hub.all.List() {
		ch <- evt
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
func (hub *Hub) Watch(kind any, done <-chan struct{}) <-chan Event {
	ch := make(chan Event)

	h := chanEventHandler{
		ch: ch,
	}

	hub.AddHandler(kind, h)
	go func() {
		<-done
		hub.RemoveHandler(kind, h)
		close(ch)
	}()

	return ch
}
func (hub *Hub) WatchAll(done <-chan struct{}) <-chan Event {
	ch := make(chan Event)

	hub.all.Add(ch)
	go func() {
		<-done
		hub.all.Remove(ch)
		close(ch)
	}()

	return datastruct.DynamicChan(ch)
}
