package events

import (
	"context"
	"time"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/bluemir/0xC0DE/internal/events/queue"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

func NewHub(ctx context.Context) (*Hub, error) {
	in := make(chan Event)
	go func() {
		<-ctx.Done()
		close(in)
	}()

	q := queue.Queue(in)

	hub := &Hub{
		ctx:      ctx,
		values:   datastruct.Map[any, any]{},
		handlers: datastruct.Map[string, datastruct.Set[Handler]]{},
		all:      datastruct.NewSet[Handler](),
		in:       in,
	}

	go func() {
		for evt := range q {
			hub.broadcast(evt)
		}
	}()

	return hub, nil
}

type Hub struct {
	ctx context.Context

	values datastruct.Map[any, any]

	handlers datastruct.Map[string, datastruct.Set[Handler]]
	all      datastruct.Set[Handler]

	in chan<- Event
}

func (h *Hub) FireEvent(kind string, detail any) {
	logrus.Tracef("fire event: %s - %#v", kind, detail)

	h.in <- Event{
		Id:     xid.New().String(),
		At:     time.Now(),
		Kind:   kind,
		Detail: detail,
	}
}
func (h *Hub) broadcast(evt Event) {
	handlers, _ := h.handlers.GetOrSet(evt.Kind, datastruct.NewSet[Handler]())

	for _, handler := range handlers.List() {
		handler.Handle(h, evt)
	}

	for _, handler := range h.all.List() {
		handler.Handle(h, evt)
	}
}
func (h *Hub) Close() {
	close(h.in)
}

func (h *Hub) AddEventHandler(kind string, handler Handler) {
	if handler == nil {
		return
	}
	set, _ := h.handlers.GetOrSet(kind, datastruct.NewSet[Handler]())

	set.Add(handler)
}
func (h *Hub) RemoveEventHandler(kind string, handler Handler) {
	set, _ := h.handlers.GetOrSet(kind, datastruct.NewSet[Handler]())

	set.Remove(handler)
}

func (h *Hub) AddAllEventHandler(handler Handler) {
	h.all.Add(handler)
}
func (h *Hub) RemoveAllEventHandler(handler Handler) {
	h.all.Remove(handler)
}
func (h *Hub) AddEventListener(kind string, l Listener) {
	h.AddEventHandler(kind, chanEventHandler{l})
}
func (h *Hub) RemoveEventListener(kind string, l Listener) {
	h.RemoveEventHandler(kind, chanEventHandler{l})
}
func (h *Hub) AddAllEventListener(l Listener) {
	h.AddAllEventHandler(chanEventHandler{l})
}
func (h *Hub) RemoveAllEventListener(l Listener) {
	h.RemoveAllEventHandler(chanEventHandler{l})
}

func (h *Hub) WatchEvent(kind string, done <-chan struct{}) <-chan Event {
	c := make(chan Event)
	h.AddEventListener(kind, c)
	go func() {
		<-done
		h.RemoveEventListener(kind, c)
		close(c)
	}()

	return c
}
func (h *Hub) WatchAllEvent(done <-chan struct{}) <-chan Event {
	c := make(chan Event)
	h.AddAllEventListener(c)
	go func() {
		<-done
		h.RemoveAllEventListener(c)
		close(c)
	}()

	return c
}

func (h *Hub) Context() context.Context {
	return h.ctx
}

func (h *Hub) Set(key any, value any) {
	h.values.Set(key, value)
}
func (h *Hub) Get(key any) (any, bool) {
	return h.values.Get(key)
}
