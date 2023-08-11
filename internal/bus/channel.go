package bus

import (
	"sync"
)

type Channel struct {
	listeners map[chan<- Event]null
	lock      sync.RWMutex
}

type null struct{}

func NewChannel() *Channel {
	return &Channel{
		listeners: map[chan<- Event]null{},
	}
}
func (ch *Channel) broadcastEvent(evt Event) {
	ch.lock.RLock()
	defer ch.lock.RUnlock()

	for l := range ch.listeners {
		l <- evt
	}
}
func (ch *Channel) AddEventListener(l chan<- Event) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	ch.listeners[l] = null{}
}
func (ch *Channel) RemoveEventListener(l chan<- Event) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	delete(ch.listeners, l)
}
