package mux

import (
	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/sirupsen/logrus"
)

type Mux[T any] struct {
	ch <-chan T

	// channels map[chan<- T]bool
	channels datastruct.Set[chan<- T]
	done     <-chan struct{}
}

func New[T any](ch <-chan T) (*Mux[T], error) {

	done := make(chan struct{})

	m := &Mux[T]{
		ch: ch,

		channels: datastruct.NewSet[chan<- T](),
		done:     done,
	}
	go func(m *Mux[T]) {
		for v := range ch {
			m.broadcast(v)
		}
		close(done) // broadcast end of mux

		logrus.Trace("termination mux broadcast")
	}(m)

	return m, nil
}

func (m *Mux[T]) Watch(done <-chan struct{}) <-chan T {
	ch := make(chan T)

	m.channels.Add(ch)

	go func() {
		select { // close one of them
		case <-done:
		case <-m.done:
		}
		// cleanup
		m.channels.Remove(ch)
		close(ch)
	}()

	return datastruct.DynamicChan(ch)
}

func (m *Mux[T]) broadcast(v T) {
	m.channels.ForEach(func(ch chan<- T) error {
		ch <- v
		return nil
	})
}
func (m *Mux[T]) Wait() {
	<-m.done
}
