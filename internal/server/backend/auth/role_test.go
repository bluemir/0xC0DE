package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestRoleCRUD(t *testing.T) {
	m := newTestManager(t)

	// Create Role
	role, err := m.CreateRole("test-admin", []auth.Rule{
		{
			Verbs: []auth.Verb{"*"},
			Selector: auth.KeyValues{
				"kind": "*",
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "test-admin", role.Name)
	assert.Len(t, role.Rules, 1)

	// Get Role
	gotRole, err := m.GetRole("test-admin")
	assert.NoError(t, err)
	assert.Equal(t, "test-admin", gotRole.Name)

	// List Role
	roles, err := m.ListRole()
	assert.NoError(t, err)
	// Might contain default admin role + test-admin
	assert.Condition(t, func() bool {
		for _, r := range roles {
			if r.Name == "test-admin" {
				return true
			}
		}
		return false
	})

	// Update Role
	gotRole.Rules = []auth.Rule{}
	err = m.UpdateRole(gotRole)
	assert.NoError(t, err)

	gotRole, err = m.GetRole("test-admin")
	assert.NoError(t, err)
	assert.Len(t, gotRole.Rules, 0)

	// Delete Role
	err = m.DeleteRole("test-admin")
	assert.NoError(t, err)

	_, err = m.GetRole("test-admin")
	assert.Error(t, err)
}
