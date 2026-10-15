package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCRUD(t *testing.T) {
	ctx := t.Context()
	m := newTestManager(t)

	// Create Group
	group, err := m.CreateGroup(ctx, "devs")
	require.NoError(t, err)
	assert.Equal(t, "devs", group.Name)

	// Create Duplicate Group
	_, err = m.CreateGroup(ctx, "devs")
	assert.Error(t, err)

	// Ensure Group (Idempotent)
	group2, err := m.EnsureGroup(ctx, "ops")
	require.NoError(t, err)
	assert.Equal(t, "ops", group2.Name)

	group3, err := m.EnsureGroup(ctx, "ops")
	require.NoError(t, err)
	assert.Equal(t, "ops", group3.Name)

	// List Group
	groups, err := m.ListGroup(ctx)
	require.NoError(t, err)

	// Check if created groups exist
	groupNames := []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}
	assert.Contains(t, groupNames, "devs")
	assert.Contains(t, groupNames, "ops")

	// Delete Group
	err = m.DeleteGroup(ctx, "devs")
	assert.NoError(t, err)

	groups, err = m.ListGroup(ctx)
	require.NoError(t, err)

	groupNames = []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}
	assert.NotContains(t, groupNames, "devs")
	assert.Contains(t, groupNames, "ops")
}
