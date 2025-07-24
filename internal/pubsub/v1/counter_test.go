package pubsub_test

import (
	"sync"

	"github.com/bluemir/0xC0DE/internal/pubsub/v1"
	"github.com/sirupsen/logrus"
)

type CounterHandler struct {
	lock  sync.RWMutex
	count int
}

func (h *CounterHandler) Handle(evt pubsub.Message) {
	h.lock.Lock()
	defer h.lock.Unlock()

	logrus.Trace("counter called")
	h.count++
}
func (h *CounterHandler) GetCount() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.count
}
