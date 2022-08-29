package events

import (
	"sync"
)

func newChannel[Event any]() *channel[Event] {
	return &channel[Event]{
		handlers: map[Handler[Event]]struct{}{},
	}
}

type channel[Event any] struct {
	handlers map[Handler[Event]]struct{}
	lock     sync.RWMutex
}

type Handler[Event any] interface {
	Handle(Event)
}

func (ch *channel[Event]) addListener(h Handler[Event]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.handlers[h] = struct{}{}
}
func (ch *channel[Event]) removeListener(h Handler[Event]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	delete(ch.handlers, h)
}
func (ch *channel[Event]) broadcastEvent(evt Event) {
	for h := range ch.handlers {
		h.Handle(evt)
	}
}
