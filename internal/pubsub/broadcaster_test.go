package pubsub_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/stretchr/testify/assert"
)

type Message struct {
	str string
}

func TestBroadcaster(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	b := pubsub.NewBroadcaster[Message]()

	go func() {
		//time.Sleep(500 * time.Millisecond)
		b.Broadcast(Message{"a"})
		b.Broadcast(Message{"b"})
		b.Close()
	}()

	{
		counter := 0
		ch := b.Watch(ctx.Done())
		for m := range ch {
			counter++
			println(m.str)
		}
		assert.Equal(t, 2, counter)
	}
}
