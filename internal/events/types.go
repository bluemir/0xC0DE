package events

import (
	"context"
	"time"
)

type Event struct {
	Id     string
	At     time.Time
	Kind   string
	Detail any
}

type Handler interface {
	Handle(ctx Context, evt Event)
}

type Listener chan<- Event

type Context interface {
	Context() context.Context

	Set(key any, value any)
	Get(key any) (any, bool)

	IHub
}

type IHub interface {
	FireEvent(kind string, detail any)

	AddEventHandler(kind string, h Handler)
	RemoveEventHandler(kind string, h Handler)

	AddAllEventHandler(h Handler)
	RemoveAllEventHandler(h Handler)

	AddEventListener(kind string, l Listener)
	RemoveEventListener(kind string, l Listener)
	AddAllEventListener(l Listener)
	RemoveAllEventListener(l Listener)

	WatchEvent(kind string, done <-chan struct{}) <-chan Event
	WatchAllEvent(done <-chan struct{}) <-chan Event
}

var _ IHub = &Hub{}
