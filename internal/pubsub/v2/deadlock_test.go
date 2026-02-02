package pubsub_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/pubsub/v2"
)

func TestHub_RecursivePublish_Deadlock(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second) // Fail fast if deadlock
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	type RecursiveEvent struct {
		Depth int
	}

	done := make(chan struct{})
	ch := hub.Watch(RecursiveEvent{}, done)
	defer close(done)

	// Wait for watcher to register
	time.Sleep(10 * time.Millisecond)

	// Trigger the first event
	go func() {
		hub.Publish(ctx, RecursiveEvent{Depth: 0})
	}()

	// Consume and republish recursively
	timeout := time.After(1 * time.Second)

	func() {
		for {
			select {
			case evt := <-ch:
				e := evt.Detail.(RecursiveEvent)
				if e.Depth >= 3 {
					return
				}
				// Recursive publish while holding the channel read loop
				// If Publish blocks waiting for this channel to read (which it is currently doing, but for the *new* event), we deadlock.
				// However, Publish is synchronous to handlers, but asynchronous to channels... wait.
				// In current Hub.Publish:
				//   hub.broadcaster.Broadcast(evt) -> sends to all channels in broadcaster
				//   Then iterates handlers.

				// Wait, Hub.Watch uses `hub.AddHandler`.
				// Let's look at Hub.Watch implementation again.
				// func (hub *Hub) Watch(kind any, done <-chan struct{}) <-chan Event {
				//    ch := make(chan Event)
				//    h := chanEventHandler{ch: ch}
				//    hub.AddHandler(kind, h)
				//    ...
				// }
				// chanEventHandler.Handle:
				// func (h chanEventHandler) Handle(ctx context.Context, evt Event) error {
				// 	 h.ch <- evt  <-- This blocks if ch is not read!
				// 	 return nil
				// }

				// So if we are HERE, we have read 'evt' from 'ch'.
				// Now we call hub.Publish.
				// hub.Publish -> iterates handlers -> finds chanEventHandler -> calls Handle -> tries to write to ch.
				// BUT 'ch' is unbuffered.
				// And we are the one reading 'ch'. We are currently busy calling Publish.
				// So we are NOT reading 'ch'.
				// So 'h.ch <- evt' will BLOCK.
				// So 'hub.Publish' will BLOCK.
				// So we (the reader) will BLOCK on 'hub.Publish'.
				// DEADLOCK.

				hub.Publish(ctx, RecursiveEvent{Depth: e.Depth + 1})
			case <-timeout:
				assert.Fail(t, "Deadlock detected or test timed out")
				return
			}
		}
	}()
}

func TestRouter_RecursivePublish_Deadlock(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx) // NOTE: Router constructor is NewRoute in current code
	require.NoError(t, err)

	done := make(chan struct{})
	// Mock NewRoute issue if needed, router.go showed NewRoute

	ch := router.Watch("test.recursive", done)
	defer close(done)
	time.Sleep(10 * time.Millisecond)

	go func() {
		router.Publish(ctx, "test.recursive", 0)
	}()

	timeout := time.After(1 * time.Second)
	func() {
		for {
			select {
			case evt := <-ch:
				depth := evt.Detail.(int)
				if depth >= 3 {
					return
				}
				router.Publish(ctx, "test.recursive", depth+1)
			case <-timeout:
				assert.Fail(t, "Deadlock detected or test timed out")
				return
			}
		}
	}()
}
