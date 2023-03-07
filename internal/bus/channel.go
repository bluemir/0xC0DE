package bus

import (
	"sync"
)

type Channel[T any] struct {
	listeners map[chan<- Event[T]]null
	lock      sync.RWMutex
}

type null struct{}

func NewChannel[T any]() *Channel[T] {
	return &Channel[T]{
		listeners: map[chan<- Event[T]]null{},
	}
}
func (ch *Channel[T]) broadcastEvent(evt Event[T]) {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	for l := range ch.listeners {
		l <- evt
	}
}
func (ch *Channel[T]) AddEventListener(l chan<- Event[T]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.listeners[l] = null{}
}
func (ch *Channel[T]) RemoveEventListener(l chan<- Event[T]) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	delete(ch.listeners, l)
}
