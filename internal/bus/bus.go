package bus

import (
	"context"
	"sync"
	"time"

	"github.com/rs/xid"
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
type Bus struct {
	q        chan<- Event
	lock     sync.RWMutex
	channels map[string]*Channel
}
type Event struct {
	Id     string
	At     time.Time
	Kind   string
	Detail any
}

func NewBus(ctx context.Context) (*Bus, error) {
	q := make(chan Event)

	b := &Bus{
		q:        q,
		channels: map[string]*Channel{},
	}

	go b.runBroadcaster(ctx, q)

	return b, nil
}
func (bus *Bus) FireEvent(kind string, detail any) {
	bus.q <- Event{
		Id:     xid.New().String(),
		At:     time.Now(),
		Kind:   kind,
		Detail: detail,
	}
}
func (bus *Bus) AddEventListener(kind string, l chan<- Event) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[kind]
	if ok {
		ch.AddEventListener(l)
		return
	}

	bus.channels[kind] = NewChannel()
	bus.channels[kind].AddEventListener(l)
}
func (bus *Bus) RemoveEventListener(kind string, l chan<- Event) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	ch, ok := bus.channels[kind]
	if !ok {
		return
	}
	ch.RemoveEventListener(l)
}
func (bus *Bus) WatchEvent(kind string, done <-chan struct{}) (<-chan Event, error) {
	c := make(chan Event)

	bus.AddEventListener(kind, c)
	go func() {
		<-done
		bus.RemoveEventListener(kind, c)
	}()
	return c, nil
}
func (bus *Bus) WatchAllEvent(done <-chan struct{}) (<-chan Event, error) {
	return nil, nil
}
func (bus *Bus) runBroadcaster(ctx context.Context, q <-chan Event) {
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
func (bus *Bus) broadcastEvent(evt Event) {
	bus.lock.RLock()
	defer bus.lock.RUnlock()

	ch, ok := bus.channels[evt.Kind]
	if ok {
		ch.broadcastEvent(evt)
	}

	bus.channels["*"].broadcastEvent(evt)
}
