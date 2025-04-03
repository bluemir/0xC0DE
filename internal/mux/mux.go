package mux

import (
	"context"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/sirupsen/logrus"
)

type Mux[T any] struct {
	ch <-chan T

	channels map[chan<- T]struct{}
	// TODO need lock?
}

func New[T any](ctx context.Context, ch <-chan T) (*Mux[T], error) {
	m := &Mux[T]{
		ch: ch,

		channels: map[chan<- T]struct{}{},
	}
	go func(m *Mux[T]) {
		defer func() {
			// close all channels
			for ch := range m.channels {
				close(ch)
			}

			logrus.Trace("termination mux broadcast")
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-ch:
				if !more {
					return
				}
				m.broadcast(v)
			}
		}
	}(m)

	return m, nil
}

func (m *Mux[T]) Watch(done <-chan struct{}) <-chan T {
	ch := make(chan T)

	m.channels[ch] = struct{}{}

	go func() {
		// cleanup
		<-done
		delete(m.channels, ch)
		close(ch)
	}()

	return datastruct.DynamicChan(ch)
}

func (m *Mux[T]) broadcast(v T) {
	for ch := range m.channels {
		ch <- v
	}
}
