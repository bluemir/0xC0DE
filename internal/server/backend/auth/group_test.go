package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCRUD(t *testing.T) {
	m := newTestManager(t)

	// Create Group
	group, err := m.CreateGroup("devs")
	require.NoError(t, err)
	assert.Equal(t, "devs", group.Name)

	// Create Duplicate Group (Should fail? Or GORM might return error)
	_, err = m.CreateGroup("devs")
	assert.Error(t, err)

	// Ensure Group (Idempotent)
	group2, err := m.EnsureGroup("ops")
	require.NoError(t, err)
	assert.Equal(t, "ops", group2.Name)

	group3, err := m.EnsureGroup("ops")
	require.NoError(t, err)
	assert.Equal(t, "ops", group3.Name)

	// List Group
	groups, err := m.ListGroup()
	require.NoError(t, err)
	// We have devs and ops.
	// Order is not guaranteed, but checks existence
	assert.Len(t, groups, 2)

	names := []string{groups[0].Name, groups[1].Name}
	assert.Contains(t, names, "devs")
	assert.Contains(t, names, "ops")

	// Delete Group
	err = m.DeleteGroup("devs")
	assert.NoError(t, err)

	groups, err = m.ListGroup()
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, "ops", groups[0].Name)
}
