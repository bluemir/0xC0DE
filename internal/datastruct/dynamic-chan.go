package datastruct

type dynamicChanConfig struct {
	callback func(int)
}

type DynamicChanOption[T any] func(*dynamicChanConfig)

func WithLengthCallback[T any](callback func(int)) DynamicChanOption[T] {
	return func(c *dynamicChanConfig) {
		c.callback = callback
	}
}

func DynamicChan[T any](in <-chan T, opts ...DynamicChanOption[T]) <-chan T {
	out := make(chan T)
	conf := &dynamicChanConfig{}
	for _, opt := range opts {
		opt(conf)
	}

	go func() {
		defer close(out)

		store := NewQueue[T]()
		notifyLen := func() {
			if conf.callback != nil {
				conf.callback(store.Len())
			}
		}

		for {
			if store.Len() == 0 {
				evt, more := <-in
				if !more {
					notifyLen() // Should be 0
					return
				}
				store.Add(evt)
				notifyLen()
				continue
			}

			select {
			case evt, more := <-in:
				if !more {
					for store.Len() > 0 {
						out <- store.Front()
						store.Pop()
						notifyLen()
					}
					return
				}
				store.Add(evt)
				notifyLen()
			case out <- store.Front():
				store.Pop()
				notifyLen()
			}
		}
	}()

	return out
}
