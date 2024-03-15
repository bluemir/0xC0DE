package events_test

import (
	"context"
	"sync"

	"github.com/bluemir/0xC0DE/internal/events"
)

type Recoder struct {
	recodes []events.Event
	lock    sync.RWMutex
}

func NewRecoder(ctx context.Context, hub *events.Hub) *Recoder {
	recoder := Recoder{}
	ch := hub.WatchAllEvent(ctx.Done())
	go recoder.run(ch)
	return &recoder
}
func (r *Recoder) run(ch <-chan events.Event) {
	for evt := range ch {
		r.lock.Lock()
		r.recodes = append(r.recodes, evt)
		r.lock.Unlock()
	}
}

func (r *Recoder) History() []string {
	ret := []string{}

	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, recode := range r.recodes {
		ret = append(ret, recode.Kind)
	}
	return ret
}
