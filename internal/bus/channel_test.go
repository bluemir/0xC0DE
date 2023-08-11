package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Message struct {
}

func TestChannel(t *testing.T) {
	channel := NewChannel()

	forTest := make(chan Event)

	channel.AddEventListener(forTest)

	assert.Equal(t, 1, len(channel.listeners))

	channel.RemoveEventListener(forTest)

	assert.Equal(t, 0, len(channel.listeners))
}
