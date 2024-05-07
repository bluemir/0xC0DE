package workers

import (
	"context"

	"github.com/bluemir/0xC0DE/internal/events/queue"
	"github.com/sirupsen/logrus"
)

type option struct {
	readBufSize  int
	writeBufSize int
	workerNum    int
}
type OptionFn func(*option)

func Run[InT any, OutT any](ctx context.Context, fn func(context.Context, InT) (OutT, error), opts ...OptionFn) (chan<- InT, <-chan OutT, <-chan error) {
	opt := option{
		readBufSize:  16,
		writeBufSize: 16,
		workerNum:    4,
	} // default

	for _, optFn := range opts {
		optFn(&opt)
	}

	ich := make(chan InT, opt.readBufSize)
	och := make(chan OutT, opt.writeBufSize)
	ech := make(chan error)

	doneCh := make(chan struct{})

	for i := 0; i < opt.workerNum; i++ {
		go worker(ctx, fn, ich, och, ech, doneCh)
	}

	go func() {
		defer close(och)
		defer close(ech)

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
				// case <- ech:
				//   if opt.exitOnError { return }
				//    send to out side? or collect and result.Errs()?
			}
		}
	}()

	return ich, och, queue.Queue(ech)

	// result.Out, result.In, result.Err, result.Add(array...), return result.Collect(), result.Errs()
}
func worker[InT any, OutT any](ctx context.Context, fn func(context.Context, InT) (OutT, error), in <-chan InT, out chan<- OutT, errc chan<- error, doneCh chan<- struct{}) {
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
func ReadBufferSize(n int) OptionFn {
	return func(opt *option) {
		opt.readBufSize = n
	}
}
func WorkerNum(n int) OptionFn {
	return func(opt *option) {
		opt.workerNum = n
	}
}
