package datastruct_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func TestSet(t *testing.T) {
	s := datastruct.NewSet[int]()

	// Add
	s.Add(1)
	s.Add(2)
	s.Add(1) // duplicate

	// Has
	assert.True(t, s.Has(1))
	assert.True(t, s.Has(2))
	assert.False(t, s.Has(3))

	// Remove
	s.Remove(1)
	assert.False(t, s.Has(1))
	assert.True(t, s.Has(2))

	// Range
	// Add more
	s.Add(3)
	s.Add(4)
	count := 0
	err := s.Range(func(item int) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	// 2, 3, 4 should be there
	assert.Equal(t, 3, count)

	// Range error case
	err = s.Range(func(item int) error {
		return errors.New("range error")
	})
	assert.Error(t, err)

	// Clear
	s.Clear()
	assert.False(t, s.Has(2))
	assert.False(t, s.Has(3))
	assert.False(t, s.Has(4))

	countAfterClear := 0
	s.Range(func(item int) error {
		countAfterClear++
		return nil
	})
	assert.Equal(t, 0, countAfterClear)
}
