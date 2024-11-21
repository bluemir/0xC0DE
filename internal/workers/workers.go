package workers

import (
	"context"

	"github.com/bluemir/0xC0DE/internal/datastruct"
	"github.com/sirupsen/logrus"
)

type option struct {
	workerNum int
}
type OptionFn func(*option)

type WorkerFunc[InType any, OutType any] func(context.Context, InType) (OutType, error)

func Run[InType any, OutType any](ctx context.Context, in <-chan InType, fn WorkerFunc[InType, OutType], opts ...OptionFn) (<-chan OutType, <-chan error) {
	opt := option{
		workerNum: 4,
	} // default

	for _, optFn := range opts {
		optFn(&opt)
	}

	inputCh := datastruct.DynamicChan(in)
	outputCh := make(chan OutType)
	errorCh := make(chan error)

	doneCh := make(chan struct{})

	for i := 0; i < opt.workerNum; i++ {
		go worker(ctx, fn, inputCh, outputCh, errorCh, doneCh)
	}

	go func() {
		defer close(outputCh)
		defer close(errorCh)

		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-doneCh:
				count++
				if count >= opt.workerNum {
					return
				}
			}
		}
	}()

	return datastruct.DynamicChan(outputCh), datastruct.DynamicChan(errorCh)

	// result.Out, result.In, result.Err, result.Add(array...), return result.Collect(), result.Errs()
}
func worker[InType any, OutType any](ctx context.Context, fn WorkerFunc[InType, OutType], in <-chan InType, out chan<- OutType, errc chan<- error, doneCh chan<- struct{}) {
	defer func() {
		doneCh <- struct{}{}
	}()

	for inV := range in {
		outV, err := fn(ctx, inV)
		if err != nil {
			select {
			case errc <- err:
			default:
				logrus.Error(err)
			}
			continue
		}

		out <- outV
	}
}

func WorkerNum(n int) OptionFn {
	return func(opt *option) {
		opt.workerNum = n
	}
}
