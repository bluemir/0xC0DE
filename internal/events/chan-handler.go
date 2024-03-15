package events

type chanEventHandler struct {
	ch chan<- Event
}

func (h chanEventHandler) Handle(ctx Context, evt Event) {
	h.ch <- evt
}
