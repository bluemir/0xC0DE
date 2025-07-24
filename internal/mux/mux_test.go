package mux_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluemir/0xC0DE/internal/mux"
	"github.com/stretchr/testify/assert"
)

type Message struct {
	str string
}

func TestMux(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan Message)
	m, err := mux.New(ch)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		//time.Sleep(500 * time.Millisecond)
		ch <- Message{"a"}
		ch <- Message{"b"}
		close(ch)
	}()

	{
		counter := 0
		ch := m.Watch(ctx.Done())
		for m := range ch {
			counter++
			println(m.str)
		}
		assert.Equal(t, 2, counter)
	}
}
