package retry

import (
	"time"

	"github.com/cockroachdb/errors"
)

type RetryOptionFn func(*RetryOption)
type RetryOption struct {
	DelayFunc func(int) time.Duration
}

func Retry(maxRetry int, fn func() error, opts ...RetryOptionFn) error {
	opt := RetryOption{
		DelayFunc: exponential,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	var err error
	//for i := 0; i < maxRetry; i++ {
	for i := range maxRetry {
		if err = fn(); err == nil {
			return nil // success
		}

		// TODO context?
		// TODO time calc?
		<-time.After(time.Duration(i*i) * time.Second)
	}
	return errors.Wrapf(err, "failed %d time.", maxRetry)
}

func exponential(try int) time.Duration {
	return time.Duration(try) * time.Second
}
func Exponential() func(int) time.Duration {
	return exponential
}
