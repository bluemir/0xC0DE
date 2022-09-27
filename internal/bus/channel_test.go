package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Message struct {
}

type HandlerForTest struct {
}

func (HandlerForTest) Handle(Event[Message]) {
}

func TestChannel(t *testing.T) {
	channel := NewChannel[Message]()

	channel.AddEventListener(HandlerForTest{})

	assert.Equal(t, 1, len(channel.listeners))

	channel.RemoveEventListener(HandlerForTest{})

	assert.Equal(t, 0, len(channel.listeners))
}
