package pubsub

import (
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(Message{})
}

type Message struct {
	Id     string
	At     time.Time
	Kind   string
	Detail any `gorm:"type:bytes;serializer:gob"`
}

type Handler interface {
	Handle(msg Message) error
}

type Listener chan<- Message

type IHub interface {
	// kind must not contain "*"
	Publish(kind string, detail any)

	// "*" in kind mean all, split by "."
	AddHandler(kind string, h Handler)
	RemoveHandler(kind string, h Handler)

	// "*" in kind mean all, split by "."
	AddListener(kind string, l Listener)
	RemoveListener(kind string, l Listener)

	// "*" in kind mean all, split by "."
	Watch(kind string, done <-chan struct{}) <-chan Message
	// eg) for evt := range hub.Watch("test", ctx.Done())
}

var _ IHub = (*Hub)(nil)
