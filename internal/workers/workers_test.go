package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/bluemir/0xC0DE/internal/workers"
	"github.com/stretchr/testify/assert"
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
	}, 5) // 10 jobs per second

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	in, out := workers.Run[int, int](ctx, func(ctx context.Context, a int) (int, error) {
		return a * a, nil
	}, 5) // 10 jobs per second

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
