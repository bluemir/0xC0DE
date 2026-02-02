package datastruct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func TestDynamicChan_WithLengthCallback(t *testing.T) {
	in := make(chan int)

	maxLen := 0
	callbackCount := 0

	out := datastruct.DynamicChan(in, datastruct.WithLengthCallback[int](func(n int) {
		if n > maxLen {
			maxLen = n
		}
		callbackCount++
	}))

	// Send 3 items without reading, queue should grow
	in <- 1
	in <- 2
	in <- 3

	// Ensure callback was called and maxLen increased
	time.Sleep(50 * time.Millisecond)

	assert.Greater(t, maxLen, 0)
	assert.Greater(t, callbackCount, 0)

	// Read everything
	<-out
	<-out
	<-out

	close(in)
	for range out {
	} // drainage
}
