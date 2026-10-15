package pubsub

import (
	"context"
	"reflect"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

// 발행은 두 가지다. 부르는 쪽이 결과를 쓰는지로 갈린다.
//
//   - Notify  알리기만 한다. 돌아올 때까지 기다리지 않고 에러도 보지 않는다
//   - Dispatch 그 자리에서 핸들러를 돌리고 에러를 모아 돌려준다
//
// 한 가지로 묶으면 둘 중 하나는 반드시 어긋난다. 게시물 알림은 SSE 로 흘려보내면
// 그만이라 응답을 기다릴 이유가 없고, 사용자 삭제에 딸린 정리는 실패를 삼키면
// 넣은 의미가 없다.
type IHub interface {
	Notify(ctx context.Context, detail any)
	Dispatch(ctx context.Context, detail any) error
	AddHandler(kind any, handler Handler[any])
	RemoveHandler(kind any, handler Handler[any])
	Watch(kind any, done <-chan struct{}, opts ...WatchOption) <-chan Event[any]
	WatchAll(done <-chan struct{}, opts ...WatchOption) <-chan Event[any]
}

type watchConfig struct {
	bufferSize int
}

type WatchOption func(*watchConfig)

func WithBuffer(n int) WatchOption {
	return func(c *watchConfig) {
		c.bufferSize = n
	}
}

var _ IHub = (*Hub)(nil)

type Hub struct {
	handlers    datastruct.Map[reflect.Type, datastruct.Set[Handler[any]]]
	broadcaster *Broadcaster[Event[any]]

	// processContext 는 Notify 로 넘긴 이벤트가 쓸 context 다.
	// 요청 context 를 그대로 쓰면 응답이 끝나는 순간 취소되어, 뒤늦게 도는 핸들러가
	// 이미 죽은 context 를 받는다. jobs.Manager 가 백그라운드 작업에 쓰는 것과 같다.
	processContext context.Context
	queue          chan<- Event[any]
}

func NewHub(ctx context.Context) (*Hub, error) {
	queue := make(chan Event[any])

	hub := &Hub{
		broadcaster:    NewBroadcaster[Event[any]](),
		processContext: ctx,
		queue:          queue,
	}

	// Notify 는 부르는 쪽을 막지 않으면서 순서는 지켜야 한다.
	// DynamicChan 이 받는 쪽에서 무한히 쌓아 주므로 보내는 쪽이 밀리지 않고,
	// 꺼내 쓰는 goroutine 이 하나뿐이라 들어온 차례대로 나간다.
	go func() {
		notifications := datastruct.DynamicChan[Event[any]](queue)

		for {
			select {
			case <-ctx.Done():
				return
			case evt, more := <-notifications:
				if !more {
					return
				}
				if err := hub.deliver(evt); err != nil {
					logrus.Warnf("event handler failed. kind: %s, id: %s: %+v", evt.Kind, evt.Id, err)
				}
			}
		}
	}()

	return hub, nil
}

type keyTypeHub struct{}

var keyHub = keyTypeHub{}

func HubFrom(ctx context.Context) *Hub {
	return ctx.Value(keyHub).(*Hub)
}

func (h chanEventHandler[T]) Handle(ctx context.Context, evt Event[T]) error {
	h.ch <- evt
	return nil
}

type chanEventHandler[T any] struct {
	ch chan<- Event[T]
}

// Notify 는 이벤트를 큐에 넣고 바로 돌아온다. 핸들러의 에러는 로그로만 남는다.
// 알림처럼 결과가 부르는 쪽의 판단에 들어가지 않는 이벤트에 쓴다.
func (hub *Hub) Notify(ctx context.Context, detail any) {
	// 큐를 닫지 않는다. 닫아 두면 종료 중에 들어온 Notify 가 닫힌 채널에 써서 죽는다.
	// 대신 꺼내는 쪽이 멈추고, 여기서는 더 넣지 않는다.
	select {
	case <-hub.processContext.Done():
	case hub.queue <- hub.newEvent(hub.processContext, detail):
	}
}

// Dispatch 는 핸들러를 그 자리에서 돌리고 에러를 모아 돌려준다.
// 처리에 실패했다는 것을 부르는 쪽이 알아야 하는 이벤트에 쓴다.
//
// 핸들러는 부르는 쪽의 트랜잭션 밖에서 돈다. 에러를 받는다고 해서
// 이미 끝난 쓰기가 되돌아가지는 않는다.
func (hub *Hub) Dispatch(ctx context.Context, detail any) error {
	return hub.deliver(hub.newEvent(ctx, detail))
}

func (hub *Hub) newEvent(ctx context.Context, detail any) Event[any] {
	return Event[any]{
		Context: context.WithValue(ctx, keyHub, hub),
		Id:      xid.New().String(),
		At:      time.Now(),
		Detail:  detail,
		Kind:    reflect.TypeOf(detail).String(),
	}
}

// deliver 는 WatchAll 로 흘려보내고 kind 가 맞는 핸들러를 모두 돌린다.
// 하나가 실패해도 나머지를 멈추지 않는다. 정리 핸들러가 여럿일 때
// 앞선 하나 때문에 뒤쪽이 통째로 빠지면 치우려던 것이 더 남는다.
func (hub *Hub) deliver(evt Event[any]) error {
	hub.broadcaster.Broadcast(evt)

	handlers, ok := hub.handlers.Get(reflect.TypeOf(evt.Detail))
	if !ok {
		return nil
	}

	snapshot := []Handler[any]{}
	if err := handlers.Range(func(handler Handler[any]) error {
		snapshot = append(snapshot, handler)
		return nil
	}); err != nil {
		logrus.Debug(err)
	}

	errs := []error{}
	for _, handler := range snapshot {
		if err := handler.Handle(evt.Context, evt); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// AddHandler 는 kind 의 타입에 핸들러를 건다.
// handler 는 해시 가능해야 한다. 핸들러 집합이 sync.Map 이라 func 값이나
// func 필드를 품은 구조체를 그대로 넣으면 panic 한다. 포인터로 감싸서 넣는다.
func (hub *Hub) AddHandler(kind any, handler Handler[any]) {
	handlers, _ := hub.handlers.GetOrSet(reflect.TypeOf(kind), datastruct.NewSet[Handler[any]]())
	handlers.Add(handler)
}
func (hub *Hub) RemoveHandler(kind any, handler Handler[any]) {
	handlers, _ := hub.handlers.GetOrSet(reflect.TypeOf(kind), datastruct.NewSet[Handler[any]]())
	handlers.Remove(handler)
}
func (hub *Hub) Watch(kind any, done <-chan struct{}, opts ...WatchOption) <-chan Event[any] {
	conf := watchConfig{
		bufferSize: -1,
	}
	for _, opt := range opts {
		opt(&conf)
	}

	var ch chan Event[any]
	if conf.bufferSize >= 0 {
		ch = make(chan Event[any], conf.bufferSize)
	} else {
		ch = make(chan Event[any])
	}

	h := chanEventHandler[any]{
		ch: ch,
	}

	hub.AddHandler(kind, h)
	go func() {
		<-done
		hub.RemoveHandler(kind, h)
		close(ch)
	}()

	if conf.bufferSize >= 0 {
		return ch
	}

	// Use DynamicChan to prevent deadlock during recursive publishing
	return datastruct.DynamicChan(ch)
}
func (hub *Hub) WatchAll(done <-chan struct{}, opts ...WatchOption) <-chan Event[any] {
	return hub.broadcaster.Watch(done, opts...)
}
