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
}
func worker[InType any, OutType any](ctx context.Context, fn WorkerFunc[InType, OutType], in <-chan InType, outCh chan<- OutType, errCh chan<- error, doneCh chan<- struct{}) {
	defer func() {
		doneCh <- struct{}{}
	}()

	for inV := range in {
		outV, err := fn(ctx, inV)
		if err != nil {
			select {
			case errCh <- err:
			default:
				logrus.Error(err)
			}
			continue
		}

		outCh <- outV
	}
}

func WorkerNum(n int) OptionFn {
	return func(opt *option) {
		opt.workerNum = n
	}
}

// RunSlice processes a slice of inputs concurrently and collects results.
// It returns collected outputs and any errors that occurred.
// This is a convenience wrapper around Run for slice-based inputs.
func RunSlice[InType any, OutType any](ctx context.Context, inputs []InType, fn WorkerFunc[InType, OutType], opts ...OptionFn) ([]OutType, []error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	in := make(chan InType)
	go func() {
		defer close(in)
		for _, input := range inputs {
			select {
			case <-ctx.Done():
				return
			case in <- input:
			}
		}
	}()

	outCh, errCh := Run(ctx, in, fn, opts...)

	var results []OutType
	var errors []error

	// Collect all results and errors
	outDone := false
	errDone := false
	for !outDone || !errDone {
		select {
		case result, ok := <-outCh:
			if !ok {
				outDone = true
				continue
			}
			results = append(results, result)
		case err, ok := <-errCh:
			if !ok {
				errDone = true
				continue
			}
			errors = append(errors, err)
		}
	}

	return results, errors
}
