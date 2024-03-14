package workers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/workers"
)

func TestSimpleWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 1)

	for i := 0; i < 30; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 30)
	assert.Equal(t, 0, response[0])
	assert.Equal(t, 9, response[3])
	assert.Equal(t, 100, response[10])
	assert.Nil(t, ctx.Err())
}
func TestMultipleWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 5)

	for i := 0; i < 30; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 30)
	assert.Nil(t, ctx.Err())
}
func TestWorkerWithLargeNumber(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 128)

	for i := 0; i < 30; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 30)
	assert.Nil(t, ctx.Err())
}
func TestWorkerWithDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		time.Sleep(500 * time.Millisecond)
		return a * a, nil
	}, 5)

	for i := 0; i < 30; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 30)
	assert.Nil(t, ctx.Err())
}
func TestManyJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 5)

	go func() {
		for i := 0; i < 128; i++ {
			in <- i
		}
		close(in)
	}() // goroutine 으로 실행 하지 않으면, read buf 가 가득차서 block 됨

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 128)
	assert.Nil(t, ctx.Err())
}
func TestOptionReadBufSize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 5, workers.ReadBufferSize(128)) // with big read buffer

	for i := 0; i < 128; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 128)
	assert.Nil(t, ctx.Err())

}

func TestErrorOnWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		if a == 13 {
			return 0, errors.New("dummy error")
		}
		return a * a, nil
	}, 5)

	for i := 0; i < 20; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 19)
	assert.Nil(t, ctx.Err())
}

func TestMultipleErrorOnWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		if a < 15 {
			return 0, errors.New("dummy error")
		}
		return a * a, nil
	}, 5)

	for i := 0; i < 20; i++ {
		in <- i
	}
	close(in)

	response := []int{}

	for ret := range out {
		response = append(response, ret)
	}

	assert.Len(t, response, 5)
	assert.Nil(t, ctx.Err())
}
