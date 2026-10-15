package pubsub_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/pubsub"
)

// blockingHandler 는 풀어 줄 때까지 Handle 에서 버틴다
type blockingHandler struct {
	release chan struct{}
	entered chan struct{}
}

func (h blockingHandler) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	h.entered <- struct{}{}
	<-h.release
	return nil
}

// failingHandler 는 늘 실패한다
type failingHandler struct {
	message string
}

func (h failingHandler) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	return errors.New(h.message)
}

// Notify 는 핸들러가 버티고 있어도 부르는 쪽을 막지 않는다
func TestHub_NotifyDoesNotBlock(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	handler := blockingHandler{
		release: make(chan struct{}),
		entered: make(chan struct{}, 1),
	}
	hub.AddHandler(EventForTest{}, handler)
	defer close(handler.release)

	returned := make(chan struct{})
	go func() {
		defer close(returned)
		// 핸들러가 첫 이벤트를 붙들고 있어도 계속 넣을 수 있어야 한다
		for i := range 5 {
			hub.Notify(ctx, EventForTest{Message: fmt.Sprintf("e%d", i)})
		}
	}()

	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("Notify blocked while a handler was busy")
	}

	// 핸들러가 실제로 붙들려 있었는지 확인한다. 아니면 위가 그냥 빨랐던 것뿐이다
	select {
	case <-handler.entered:
	case <-time.After(time.Second):
		t.Fatal("handler never ran")
	}
}

// Notify 로 넣은 순서대로 핸들러에 닿는다
func TestHub_NotifyKeepsOrder(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	const count = 50

	received := make(chan string, count)
	hub.AddHandler(EventForTest{}, handlerFunc(func(ctx context.Context, evt pubsub.Event[any]) error {
		received <- evt.Detail.(EventForTest).Message
		return nil
	}))

	for i := range count {
		hub.Notify(ctx, EventForTest{Message: fmt.Sprintf("e%d", i)})
	}

	for i := range count {
		select {
		case got := <-received:
			assert.Equal(t, fmt.Sprintf("e%d", i), got, "event %d out of order", i)
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}
}

// Notify 는 핸들러 에러를 부르는 쪽으로 올리지 않는다. 그래서 반환값도 없다
func TestHub_NotifySwallowsError(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	ran := make(chan struct{}, 1)
	hub.AddHandler(EventForTest{}, handlerFunc(func(ctx context.Context, evt pubsub.Event[any]) error {
		ran <- struct{}{}
		return errors.New("boom")
	}))

	hub.Notify(ctx, EventForTest{Message: "x"})

	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("handler never ran")
	}
}

// Dispatch 는 핸들러 에러를 모아 돌려준다
func TestHub_DispatchReturnsError(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	hub.AddHandler(EventForTest{}, failingHandler{message: "first failed"})

	err = hub.Dispatch(ctx, EventForTest{Message: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first failed")
}

// 핸들러 하나가 실패해도 나머지는 돈다. 치우려던 것이 더 남으면 안 된다
func TestHub_DispatchRunsEveryHandler(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	mu := sync.Mutex{}
	ran := []string{}
	record := func(name string) *funcHandler {
		return handlerFunc(func(ctx context.Context, evt pubsub.Event[any]) error {
			mu.Lock()
			defer mu.Unlock()
			ran = append(ran, name)
			return nil
		})
	}

	hub.AddHandler(EventForTest{}, record("before"))
	hub.AddHandler(EventForTest{}, failingHandler{message: "middle failed"})
	hub.AddHandler(EventForTest{}, record("after"))

	err = hub.Dispatch(ctx, EventForTest{Message: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "middle failed")

	mu.Lock()
	defer mu.Unlock()
	assert.ElementsMatch(t, []string{"before", "after"}, ran)
}

// 핸들러가 모두 성공하면 nil 이다
func TestHub_DispatchWithoutHandler(t *testing.T) {
	ctx, cancel := testContext(t, 2*time.Second)
	defer cancel()

	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	assert.NoError(t, hub.Dispatch(ctx, EventForTest{Message: "nobody listens"}))
}

// funcHandler 는 함수 하나를 핸들러로 쓴다.
// 포인터로 등록한다. datastruct.Set 이 sync.Map 이라 키가 해시 가능해야 하는데
// func 값은 해시할 수 없다.
type funcHandler struct {
	fn func(ctx context.Context, evt pubsub.Event[any]) error
}

func handlerFunc(fn func(ctx context.Context, evt pubsub.Event[any]) error) *funcHandler {
	return &funcHandler{fn: fn}
}

func (h *funcHandler) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	return h.fn(ctx, evt)
}
