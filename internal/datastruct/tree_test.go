package datastruct_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/datastruct"
)

func TestTree(t *testing.T) {
	tree := datastruct.NewTree[string, string]()

	// Set
	tree.Set([]string{"a", "b"}, "value1")
	tree.Set([]string{"a", "c"}, "value2")

	// Get
	val, ok := tree.Get("a", "b")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	val, ok = tree.Get("a", "c")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)

	val, ok = tree.Get("a")
	assert.False(t, ok) // only children, no value at "a"

	tree.Set([]string{"a"}, "value_root_a")
	val, ok = tree.Get("a")
	assert.True(t, ok)
	assert.Equal(t, "value_root_a", val)

	// GetOrSet
	// Existing
	val, ok = tree.GetOrSet([]string{"a", "b"}, "new_value")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// New
	val, ok = tree.GetOrSet([]string{"x", "y"}, "value3")
	assert.False(t, ok)
	assert.Equal(t, "value3", val)

	val, ok = tree.Get("x", "y")
	assert.True(t, ok)
	assert.Equal(t, "value3", val)
}
