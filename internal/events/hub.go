package events

import (
	"context"
	"time"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

func NewHub(ctx context.Context) (*Hub, error) {
	return &Hub{
		ctx:      ctx,
		values:   datastruct.Map[any, any]{},
		handlers: datastruct.Map[string, datastruct.Set[Handler]]{},
		all:      datastruct.NewSet[Handler](),
	}, nil

}

type Hub struct {
	ctx context.Context

	values datastruct.Map[any, any]

	handlers datastruct.Map[string, datastruct.Set[Handler]]
	all      datastruct.Set[Handler]
}

func (h *Hub) FireEvent(kind string, detail any) {
	logrus.Tracef("fire event: %s - %#v", kind, detail)

	handlers, _ := h.handlers.GetOrSet(kind, datastruct.NewSet[Handler]())

	for _, handler := range handlers.List() {
		handler.Handle(h, Event{
			Id:     xid.New().String(),
			At:     time.Now(),
			Kind:   kind,
			Detail: detail,
		})
	}

	for _, handler := range h.all.List() {
		handler.Handle(h, Event{
			Id:     xid.New().String(),
			At:     time.Now(),
			Kind:   kind,
			Detail: detail,
		})
	}

	// TODO
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

func (h *Hub) WatchEvent(kind string, done <-chan struct{}) <-chan Event {
	c := make(chan Event)
	h.AddEventHandler(kind, ChanEventHandler{c})
	go func() {
		<-done
		h.RemoveEventHandler(kind, ChanEventHandler{c})
		close(c)
	}()

	return c
}
func (h *Hub) WatchAllEvent(done <-chan struct{}) <-chan Event {
	c := make(chan Event)
	h.AddAllEventHandler(ChanEventHandler{c})
	go func() {
		<-done
		h.RemoveAllEventHandler(ChanEventHandler{c})
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
