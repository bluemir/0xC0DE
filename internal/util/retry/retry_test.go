package retry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/util/retry"
)

func TestRetry(t *testing.T) {
	// Success on first try
	count := 0
	err := retry.Retry(3, func() error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Success on second try
	count = 0
	err = retry.Retry(3, func() error {
		count++
		if count < 2 {
			return errors.New("fail")
		}
		return nil
	}, func(o *retry.RetryOption) {
		o.DelayFunc = func(i int) time.Duration { return 0 }
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Fail all retries
	count = 0
	err = retry.Retry(3, func() error {
		count++
		return errors.New("always fail")
	}, func(o *retry.RetryOption) {
		o.DelayFunc = func(i int) time.Duration { return 0 }
	})
	assert.Error(t, err)
	assert.Equal(t, 3, count)
}

func TestExponential(t *testing.T) {
	f := retry.Exponential()
	assert.Equal(t, 1*time.Second, f(1))
	assert.Equal(t, 2*time.Second, f(2))
}
