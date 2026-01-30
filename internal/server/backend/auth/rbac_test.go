package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestRBAC(t *testing.T) {
	m := newTestManager(t)

	// Setup: User, Group, Roles
	_, err := m.CreateUser("alice", auth.WithGroup("engineers"))
	require.NoError(t, err)

	_, err = m.EnsureGroup("engineers")
	require.NoError(t, err)

	// Role: view-all
	_, err = m.CreateRole("view-all", []auth.Rule{
		{
			Verbs: []auth.Verb{"get", "list"},
			Selector: auth.KeyValues{
				"kind": "*",
			},
		},
	})
	require.NoError(t, err)

	// Role: admin-server
	_, err = m.CreateRole("admin-server", []auth.Rule{
		{
			Verbs: []auth.Verb{},
			Selector: auth.KeyValues{
				"kind": "server",
			},
		},
	})
	require.NoError(t, err)

	// Assign Role to User
	err = m.AssignRole(auth.Subject{Kind: auth.KindUser, Name: "alice"}, "view-all")
	require.NoError(t, err)

	// Assign Role to Group
	err = m.AssignRole(auth.Subject{Kind: auth.KindGroup, Name: "engineers"}, "admin-server")
	require.NoError(t, err)

	// Retrieve User for checking permissions
	alice, err := m.GetUser("alice")
	require.NoError(t, err)

	// Check Permission: User specific role
	// Alice has 'view-all', so can 'get' 'any'
	assert.True(t, m.Can(alice, "get", auth.Resource{"kind": "whatever"}))
	assert.True(t, m.Can(alice, "list", auth.Resource{"kind": "pod"}))

	// Should NOT be able to 'delete' (view-all only get/list)
	assert.False(t, m.Can(alice, "delete", auth.Resource{"kind": "pod"}))

	// Check Permission: Group inherited role
	// Engineers have 'admin-server', so can 'delete' 'server'
	// Alice is in Engineers group.
	assert.True(t, m.Can(alice, "delete", auth.Resource{"kind": "server"}))

	// Should NOT be able to 'delete' 'pod' via group (admin-server only for server)
	assert.False(t, m.Can(alice, "delete", auth.Resource{"kind": "pod"}))

	// Discard Role
	err = m.DiscardRole(auth.Subject{Kind: auth.KindUser, Name: "alice"}, "view-all")
	require.NoError(t, err)

	// Now Alice cannot 'get' 'pod' (removed user role)
	// But can still 'get' 'server' (via group role which allows *)
	assert.False(t, m.Can(alice, "get", auth.Resource{"kind": "pod"}))
	assert.True(t, m.Can(alice, "get", auth.Resource{"kind": "server"}))
}
