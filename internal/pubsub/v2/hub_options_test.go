package pubsub_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/pubsub/v2"
)

func TestHub_Watch_WithOptions(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	t.Run("Default (Dynamic)", func(t *testing.T) {
		done := make(chan struct{})
		// Default should be unbuffered (connected to pump)
		ch := hub.Watch(EventForTest{}, done)
		assert.Equal(t, 0, cap(ch))
		close(done)
	})

	t.Run("WithBuffer(10)", func(t *testing.T) {
		done := make(chan struct{})
		defer close(done)

		ch := hub.Watch(EventForTest{}, done, pubsub.WithBuffer(10))
		assert.Equal(t, 10, cap(ch)) // Must receive channel with cap 10
	})

	t.Run("WithBuffer(0)", func(t *testing.T) {
		done := make(chan struct{})
		defer close(done)

		ch := hub.Watch(EventForTest{}, done, pubsub.WithBuffer(0))
		assert.Equal(t, 0, cap(ch))
	})

	t.Run("WithBuffer(10) blocking check", func(t *testing.T) {
		// If we fill buffer, next publish might block if it was synchronous writing to channel.
		// But Hub.Publish -> Handler.Handle -> ch <- val
		// If ch is full, Handle blocks.
		// Hub.Publish iterates handlers synchronously. So Hub.Publish should block!

		hub, _ := pubsub.NewHub(ctx)
		done := make(chan struct{})
		defer close(done)

		// buffer 1
		ch := hub.Watch(EventForTest{}, done, pubsub.WithBuffer(1))

		// 1. Publish 1st event - should succeed (buffer 1)
		doneCh := make(chan struct{})
		go func() {
			hub.Publish(ctx, EventForTest{Message: "1"})
			close(doneCh)
		}()

		select {
		case <-doneCh:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("1st publish should not block")
		}

		// 2. Publish 2nd event - should block (buffer full)
		doneCh2 := make(chan struct{})
		go func() {
			hub.Publish(ctx, EventForTest{Message: "2"})
			close(doneCh2)
		}()

		select {
		case <-doneCh2:
			t.Fatal("2nd publish should block because buffer is full")
		case <-time.After(50 * time.Millisecond):
			// Expected to timeout (block)
		}

		// 3. Read one, allowing 2nd publish to proceed
		<-ch // read "1"

		select {
		case <-doneCh2:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("2nd publish should unblock after read")
		}
	})
}
