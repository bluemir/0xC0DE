package datastruct

func DynamicChan[T any](in <-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		store := NewQueue[T]()
		for {
			if store.Len() == 0 {
				evt, more := <-in
				if !more {
					return
				}
				store.Add(evt)
				continue
			}

			select {
			case evt, more := <-in:
				if !more {
					for store.Len() > 0 {
						out <- store.Front()
						store.Pop()
					}
					return
				}
				store.Add(evt)
			case out <- store.Front():
				store.Pop()
			}
		}
	}()

	return out
}
