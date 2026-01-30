package datastruct_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func TestDynamicChan(t *testing.T) {
	in := make(chan int)
	out := datastruct.DynamicChan(in)

	go func() {
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	timeout := time.After(1 * time.Second)

	received := []int{}
loop:
	for {
		select {
		case v, ok := <-out:
			if !ok {
				break loop
			}
			received = append(received, v)
		case <-timeout:
			t.Fatal("timeout")
		}
	}

	assert.Equal(t, []int{1, 2, 3}, received)
}

func TestDynamicChan_Blocking(t *testing.T) {
	in := make(chan int)
	out := datastruct.DynamicChan(in)

	// Send without receiver ready on `out`
	// DynamicChan should buffer it
	in <- 1
	in <- 2

	// Now receive
	select {
	case v := <-out:
		assert.Equal(t, 1, v)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout 1")
	}

	select {
	case v := <-out:
		assert.Equal(t, 2, v)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout 2")
	}
}
