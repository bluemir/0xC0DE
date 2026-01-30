package pubsub

import (
	"context"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

type IRouter interface {
	Publish(ctx context.Context, kind string, detail any)
	AddHandler(kind string, handler Handler)
	RemoveHandler(kind string, handler Handler)
	Watch(kind string, done <-chan struct{}) <-chan Event
	WatchAll(done <-chan struct{}) <-chan Event
}

var _ IRouter = (*Router)(nil)

type Router struct {
	handlers    datastruct.Tree[string, datastruct.Set[Handler]]
	broadcaster *Broadcaster[Event]
}

const Separator = "." // QUESTION make configurable?

func NewRoute(ctx context.Context) (*Router, error) {
	return &Router{
		broadcaster: NewBroadcaster[Event](),
	}, nil
}

type keyTypeRouter struct{}

var keyRouter = keyTypeRouter{}

func RouterFrom(ctx context.Context) *Router {
	return ctx.Value(keyRouter).(*Router)
}

func (router *Router) Publish(ctx context.Context, kind string, detail any) {
	keys := strings.Split(kind, Separator)
	handlers, ok := router.handlers.Get(keys...)

	ctx = context.WithValue(ctx, keyRouter, router)
	evt := Event{
		Context: ctx,
		Id:      xid.New().String(),
		At:      time.Now(),
		Detail:  detail,
		Kind:    kind,
	}

	if ok {
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

	router.broadcaster.Broadcast(evt)
}
func (router *Router) AddHandler(kind string, handler Handler) {
	keys := strings.Split(kind, Separator)
	handlers, _ := router.handlers.GetOrSet(keys, datastruct.NewSet[Handler]())

	handlers.Add(handler)
}
func (router *Router) RemoveHandler(kind string, handler Handler) {
	keys := strings.Split(kind, Separator)
	handlers, _ := router.handlers.GetOrSet(keys, datastruct.NewSet[Handler]())

	handlers.Remove(handler)
}
func (router *Router) Watch(kind string, done <-chan struct{}) <-chan Event {
	ch := make(chan Event)

	h := chanEventHandler{
		ch: ch,
	}

	router.AddHandler(kind, h)
	go func() {
		<-done
		router.RemoveHandler(kind, h)
		close(ch)
	}()

	return ch
}
func (router *Router) WatchAll(done <-chan struct{}) <-chan Event {
	return router.broadcaster.Watch(done)
}
