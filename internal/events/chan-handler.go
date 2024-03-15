package events

type ChanEventHandler struct {
	ch chan<- Event
}

func (h ChanEventHandler) Handle(ctx Context, evt Event) {
	h.ch <- evt
}
