package workers

import (
	"context"

	"github.com/sirupsen/logrus"
)

const bufSize = 64

// TODO Add test
func Run[InT any, OutT any](ctx context.Context, fn func(context.Context, InT) (OutT, error), workerNum int) (chan<- InT, <-chan OutT) {
	ich := make(chan InT, bufSize)
	och := make(chan OutT, bufSize)

	doneCh := make(chan struct{})

	for i := 0; i < workerNum; i++ {
		go worker(ctx, fn, ich, och, doneCh)
	}

	go func() {
		defer close(och)
		defer close(doneCh)

		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-doneCh:
				count++
				if count >= workerNum {
					return
				}
			}
		}
	}()

	return ich, och
}
func worker[InT any, OutT any](ctx context.Context, fn func(context.Context, InT) (OutT, error), in <-chan InT, out chan<- OutT, doneCh chan<- struct{}) {
	for inV := range in {
		outV, err := fn(ctx, inV)
		if err != nil {
			// TODO option? log?
			logrus.Trace(err)
			return
		}

		out <- outV
	}

	doneCh <- struct{}{}
}
