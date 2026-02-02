package pubsub_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/pubsub"
)

func TestRouter_BasicFlow(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx)
	require.NoError(t, err)

	recoder := NewRecordingHandler()

	// Test AddHandler
	router.AddHandler("foo.bar", recoder)

	// Test Publish matching
	router.Publish(ctx, "foo.bar", EventForTest{Message: "match"})
	assert.Equal(t, 1, recoder.Count())
	assert.Equal(t, "match", recoder.Events[0].Detail.(EventForTest).Message)
	assert.Equal(t, "foo.bar", recoder.Events[0].Kind)

	// Test Publish non-matching
	router.Publish(ctx, "foo.baz", EventForTest{Message: "no-match"})
	assert.Equal(t, 1, recoder.Count()) // Count should not increase
}

func TestRouter_HeirarchicalRouting(t *testing.T) {
	// Assuming the router implementation supports hierarchical routing?
	// Looking at router.go:
	// handlers.Get(keys...) where keys = strings.Split(kind, Separator)
	// It seems it does exact matching based on split keys in a tree structure.
	// But let's verify if "foo" handler receives "foo.bar" events.
	// The current implementation of Publish:
	// keys := strings.Split(kind, Separator)
	// handlers, ok := router.handlers.Get(keys...)
	// It seems it retrieves handlers explicitly at that leaf/node.
	// It does NOT seem to traverse up or down to find other handlers.
	// checking Get doc or implementation would be ideal but based on typical Tree implementation:
	// If I register at "a.b", and publish to "a.b", I get it.
	// If I register at "a", and publish to "a.b", do I get it?

	// Let's test the behavior as implemented.

	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx)
	require.NoError(t, err)

	recoderA := NewRecordingHandler()
	recoderAB := NewRecordingHandler()

	router.AddHandler("a", recoderA)
	router.AddHandler("a.b", recoderAB)

	router.Publish(ctx, "a", EventForTest{Message: "msg-a"})
	assert.Equal(t, 1, recoderA.Count())
	assert.Equal(t, 0, recoderAB.Count())

	router.Publish(ctx, "a.b", EventForTest{Message: "msg-ab"})
	assert.Equal(t, 1, recoderA.Count()) // "a" handler shouldn't receive "a.b" unless logic supports it.
	// Re-reading logic: router.handlers.Get(keys...)
	// If it returns the node's value, it's exact match or specific path match.
	// If datastruct.Tree.Get returns partial matches? Unlikely.
	// Let's assume EXACT matching for now based on code: `handlers, ok := router.handlers.Get(keys...)`
	assert.Equal(t, 1, recoderAB.Count())
}

func TestRouter_RemoveHandler(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx)
	require.NoError(t, err)

	recoder := NewRecordingHandler()
	router.AddHandler("foo", recoder)

	router.Publish(ctx, "foo", EventForTest{Message: "1"})
	assert.Equal(t, 1, recoder.Count())

	router.RemoveHandler("foo", recoder)

	router.Publish(ctx, "foo", EventForTest{Message: "2"})
	assert.Equal(t, 1, recoder.Count())
}

func TestRouter_Watch(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx)
	require.NoError(t, err)

	done := make(chan struct{})
	ch := router.Watch("foo.bar", done)

	go func() {
		time.Sleep(10 * time.Millisecond)
		router.Publish(ctx, "foo.bar", EventForTest{Message: "hello"})
	}()

	select {
	case evt := <-ch:
		assert.Equal(t, "foo.bar", evt.Kind)
		assert.Equal(t, "hello", evt.Detail.(EventForTest).Message)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for Watch")
	}

	close(done)
	select {
	case _, ok := <-ch:
		assert.False(t, ok)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for Watch close")
	}
}

func TestRouter_WatchAll(t *testing.T) {
	ctx, cancel := testContext(t, 1*time.Second)
	defer cancel()

	router, err := pubsub.NewRoute(ctx)
	require.NoError(t, err)

	done := make(chan struct{})
	ch := router.WatchAll(done)

	go func() {
		time.Sleep(10 * time.Millisecond)
		router.Publish(ctx, "any.thing", EventForTest{Message: "hello"})
	}()

	select {
	case evt := <-ch:
		assert.Equal(t, "any.thing", evt.Kind)
		assert.Equal(t, "hello", evt.Detail.(EventForTest).Message)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for WatchAll")
	}

	close(done)
	select {
	case _, ok := <-ch:
		assert.False(t, ok)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for WatchAll close")
	}
}

func TestRouter_RouterFrom(t *testing.T) {
	ctx := context.Background()
	router := &pubsub.Router{}

	// This relies on internal implementation of context key which is private.
	// We can tested via Publish which puts it in context.
	// But we cannot easily inject it without using `Publish`'s internal logic.
	// However, `router.Publish` calls handler with a context containing the router.

	router, _ = pubsub.NewRoute(ctx)

	called := false
	handler := &FunctionalHandler{
		Fn: func(c context.Context, e pubsub.Event) error {
			r := pubsub.RouterFrom(c)
			assert.Equal(t, router, r)
			called = true
			return nil
		},
	}

	router.AddHandler("test", handler)
	router.Publish(ctx, "test", nil)
	assert.True(t, called)
}

type FunctionalHandler struct {
	Fn func(context.Context, pubsub.Event) error
}

func (h *FunctionalHandler) Handle(ctx context.Context, evt pubsub.Event) error {
	return h.Fn(ctx, evt)
}
