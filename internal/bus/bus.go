package bus

import (
	"context"
	"sync"
)

/*
bus -+- channel
     |
	 +- channel
	 |
	 +- channel -+- handler
	             |
				 +- handler

pass action to bus no.... diffrent action in same event. like move

action do something..
action has all information for action
*/
type Bus[T any] struct {
	q        chan<- Event[T]
	lock     sync.RWMutex
	channels map[string]*Channel[T]
}
type Event[T any] struct {
	Kind    string
	Context T
	Detail  interface{}
}
type EventListerner[T any] interface {
	Handle(Event[T])
}

func NewBus[T any](ctx context.Context) (*Bus[T], error) {
	q := make(chan Event[T])

	b := &Bus[T]{
		q:        q,
		channels: map[string]*Channel[T]{},
	}

	go b.runBroadcaster(ctx, q)

	return b, nil
}
func (bus *Bus[T]) FireEvent(evt Event[T]) {
	bus.q <- evt
}
func (bus *Bus[T]) AddEventListener(kind string, l EventListerner[T]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[kind]
	if ok {
		ch.AddEventListener(l)
		return
	}

	bus.channels[kind] = NewChannel[T]()
	bus.channels[kind].AddEventListener(l)
}
func (bus *Bus[T]) RemoveEventListener(kind string, l EventListerner[T]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[kind]
	if !ok {
		return
	}
	ch.RemoveEventListener(l)
}
func (bus *Bus[T]) runBroadcaster(ctx context.Context, q <-chan Event[T]) {
	for {
		select {
		case evt := <-q:
			// broadcast Event
			bus.broadcastEvent(evt)
		case <-ctx.Done():
			return
		}
	}
}
func (bus *Bus[T]) broadcastEvent(evt Event[T]) {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	ch, ok := bus.channels[evt.Kind]
	if ok {
		ch.broadcastEvent(evt)
	}

	bus.channels["*"].broadcastEvent(evt)
}
