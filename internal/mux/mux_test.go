package mux_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluemir/0xC0DE/internal/mux"
	"github.com/stretchr/testify/assert"
)

type Message struct {
}

func TestMux(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan Message)
	m, err := mux.New(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}

	counter := 0

	go func(counter *int) {
		ch := m.Watch(ctx.Done())
		for range ch {
			*counter++
		}
	}(&counter)

	ch <- Message{}
	ch <- Message{}

	close(ch)

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 2, counter)
}
