package util_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/bluemir/0xC0DE/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestMergeErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	// Test merging nil
	assert.NoError(t, util.MergeErrors(nil, nil))

	// Test merging single error
	err := util.MergeErrors(err1, nil)
	assert.Error(t, err)
	// access causes directly if possible, but MultipleError is exported
	// However, MergeErrors returns `error` interface.

	// Test merging multiple
	multiErr := util.MergeErrors(err1, err2)
	assert.Error(t, multiErr)
	assert.Contains(t, multiErr.Error(), "error 1")
	assert.Contains(t, multiErr.Error(), "error 2")
}

func TestHash(t *testing.T) {
	h1 := util.Hash("hello")
	h2 := util.Hash("hello")
	h3 := util.Hash("world")

	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.NotEmpty(t, h1)
}

func TestRandomString(t *testing.T) {
	s1 := util.RandomString(10)
	s2 := util.RandomString(10)

	assert.Len(t, s1, 10)
	assert.Len(t, s2, 10)
	assert.NotEqual(t, s1, s2)

	// Check content
	for _, c := range s1 {
		// alphanumeric check
		if !strings.ContainsAny(string(c), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			t.Errorf("unexpected character: %c", c)
		}
	}
}
