package workers

import (
	"context"

	"github.com/sirupsen/logrus"
)

type option struct {
	readBufSize  int
	writeBufSize int
}
type OptionFn func(*option)

func Run[InT any, OutT any](ctx context.Context, fn func(context.Context, InT) (OutT, error), workerNum int, opts ...OptionFn) (chan<- InT, <-chan OutT) {
	opt := option{
		readBufSize:  16,
		writeBufSize: 16,
	} // default

	for _, optFn := range opts {
		optFn(&opt)
	}

	ich := make(chan InT, opt.readBufSize)
	och := make(chan OutT, opt.writeBufSize)

	doneCh := make(chan struct{})

	for i := 0; i < workerNum; i++ {
		go worker(ctx, fn, ich, och, doneCh)
	}

	go func() {
		defer close(och)

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
	defer func() {
		doneCh <- struct{}{}
	}()

	for inV := range in {
		outV, err := fn(ctx, inV)
		if err != nil {
			// TODO option? log? err ch?
			logrus.Trace(err)
			continue
		}

		out <- outV
	}
}
func ReadBufferSize(n int) OptionFn {
	return func(opt *option) {
		opt.readBufSize = n
	}
}
