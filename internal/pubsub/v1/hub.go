package pubsub

import (
	"context"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func NewHub(ctx context.Context) (*Hub, error) {
	in := make(chan Message)
	go func() {
		<-ctx.Done()
		close(in)
	}()

	q := datastruct.DynamicChan(in)

	hub := &Hub{
		ctx:      ctx,
		values:   datastruct.Map[any, any]{},
		handlers: datastruct.NewTree[string, datastruct.Set[Handler]](),
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
	ctx      context.Context
	values   datastruct.Map[any, any]
	handlers *datastruct.Tree[string, datastruct.Set[Handler]]

	in chan<- Message
}

func (h *Hub) Publish(kind string, detail any) {
	logrus.Tracef("fire event: %s - %#v", kind, detail)
	if strings.Contains(kind, "*") {
		return // error or send error as messagez?
	}

	h.in <- Message{
		Id:     xid.New().String(),
		At:     time.Now(),
		Kind:   kind,
		Detail: detail,
	}
}
func (h *Hub) broadcast(evt Message) {
	logrus.Tracef("broadcast event: %s - %#v", evt.Kind, evt.Detail)

	keys := strings.Split(evt.Kind, ".")
	{
		handlers, _ := h.handlers.GetOrSet(keys, datastruct.NewSet[Handler]())

		logrus.Tracef("%+v", &handlers)

		handlers.ForEach(func(handler Handler) error {
			handler.Handle(evt)
			return nil
		})
	}
	{
		//handler star
		for i := len(keys); i > 0; i-- {
			keys[i-1] = "*"
			handlers, _ := h.handlers.GetOrSet(keys[:i], datastruct.NewSet[Handler]())

			handlers.ForEach(func(handler Handler) error {
				handler.Handle(evt)
				return nil
			})
		}
	}
}
func (h *Hub) Close() {
	close(h.in)
}

func (h *Hub) AddHandler(kind string, handler Handler) {
	if handler == nil {
		return
	}
	keys := strings.Split(kind, ".")
	set, _ := h.handlers.GetOrSet(keys, datastruct.NewSet[Handler]())

	set.Add(handler)
}
func (h *Hub) RemoveHandler(kind string, handler Handler) {
	keys := strings.Split(kind, ".")
	set, _ := h.handlers.GetOrSet(keys, datastruct.NewSet[Handler]())

	set.Remove(handler)
}

func (h *Hub) AddListener(kind string, l Listener) {
	h.AddHandler(kind, chanEventHandler{l})
}
func (h *Hub) RemoveListener(kind string, l Listener) {
	h.RemoveHandler(kind, chanEventHandler{l})
}

func (h *Hub) Watch(kind string, done <-chan struct{}) <-chan Message {
	c := make(chan Message)
	h.AddListener(kind, c)
	go func() {
		<-done
		h.RemoveListener(kind, c)
		close(c)
	}()

	return c
}
