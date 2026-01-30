package datastruct_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func TestMap(t *testing.T) {
	m := datastruct.Map[string, int]{}

	// Test Set and Get
	m.Set("a", 1)
	val, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	val, ok = m.Get("b")
	assert.False(t, ok)
	assert.Equal(t, 0, val)

	// Test GetOrSet
	val, ok = m.GetOrSet("a", 2)
	assert.True(t, ok) // already exists
	assert.Equal(t, 1, val)

	val, ok = m.GetOrSet("c", 3)
	assert.False(t, ok) // newly set
	assert.Equal(t, 3, val)

	val, ok = m.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	// Test Range
	m.Set("d", 4)
	expected := map[string]int{
		"a": 1,
		"c": 3,
		"d": 4,
	}
	itemCount := 0
	m.Range(func(k string, v int) bool {
		itemCount++
		assert.Equal(t, expected[k], v)
		return true
	})
	assert.Equal(t, 3, itemCount)
}
