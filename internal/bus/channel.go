package bus

import (
	"sync"
)

type Channel[T any] struct {
	listeners map[EventListerner[T]]null
	lock      sync.RWMutex
}

type null struct{}

func NewChannel[T any]() *Channel[T] {
	return &Channel[T]{
		listeners: map[EventListerner[T]]null{},
	}
}
func (ch *Channel[T]) broadcastEvent(evt Event[T]) {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	for l := range ch.listeners {
		l.Handle(evt)
	}
}
func (ch *Channel[T]) AddEventListener(l EventListerner[T]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.listeners[l] = null{}
}
func (ch *Channel[T]) RemoveEventListener(l EventListerner[T]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	delete(ch.listeners, l)
}