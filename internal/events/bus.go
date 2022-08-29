package events

import (
	"context"
	"sync"
)

type none struct{}

type Bus[Event any] struct {
	q        chan<- eventWapper[Event]
	lock     sync.RWMutex
	channels map[string]*channel[Event]
	siblings map[Buslike[Event]]none
}
type Buslike[Event any] interface {
	FireEvent(name string, event Event)
}
type eventWapper[Event any] struct {
	Name  string
	Event Event
}

func New[Event any](ctx context.Context) (*Bus[Event], error) {
	q := make(chan eventWapper[Event])

	bus := &Bus[Event]{
		q:        q,
		channels: map[string]*channel[Event]{},
		siblings: map[Buslike[Event]]none{},
	}

	go bus.runBroadcaster(ctx, q)

	return bus, nil
}
func (bus *Bus[Event]) AddListener(eventName string, handler Handler[Event]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[eventName]
	if ok {
		ch.addListener(handler)
	}

	// make new one
	ch = newChannel[Event]()
	bus.channels[eventName] = ch
	ch.addListener(handler)
}
func (bus *Bus[Event]) RemoveListener(eventName string, handler Handler[Event]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[eventName]
	if ok {
		ch.removeListener(handler)
	}
}
func (bus *Bus[Event]) FireEvent(eventName string, evt Event) {
	bus.q <- eventWapper[Event]{
		Name:  eventName,
		Event: evt,
	}
}
func (bus *Bus[Event]) Attach(s Buslike[Event]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	bus.siblings[s] = none{}
}
func (bus *Bus[Event]) Detach(s Buslike[Event]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	delete(bus.siblings, s)
}

func (bus *Bus[Event]) broadcastEvent(eventName string, evt Event) {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	ch, ok := bus.channels[eventName]
	if ok {
		ch.broadcastEvent(evt)
	}

	for s := range bus.siblings {
		s.FireEvent(eventName, evt)
	}
}
func (bus *Bus[Event]) runBroadcaster(ctx context.Context, q <-chan eventWapper[Event]) {
	for {
		select {
		case data := <-q:
			// broadcast Event
			bus.broadcastEvent(data.Name, data.Event)
		case <-ctx.Done():
			return
		}
	}
}
