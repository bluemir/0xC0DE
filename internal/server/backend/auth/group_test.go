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

	// Create Duplicate Group
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

	// Check if created groups exist
	groupNames := []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}
	assert.Contains(t, groupNames, "devs")
	assert.Contains(t, groupNames, "ops")

	// Delete Group
	err = m.DeleteGroup("devs")
	assert.NoError(t, err)

	groups, err = m.ListGroup()
	require.NoError(t, err)

	groupNames = []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}
	assert.NotContains(t, groupNames, "devs")
	assert.Contains(t, groupNames, "ops")
}
