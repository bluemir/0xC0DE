package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestUserUpdateUser(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}

	user, err := m.CreateUser("bluemir", auth.WithGroup("user"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "bluemir",
	}))

	assert.Equal(t, false, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "redmir",
	}))
}
func TestAdminUpdateUser(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}

	user, err := m.CreateUser("bluemir", auth.WithGroup("admin"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, user.Groups, 1)

	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "bluemir",
	}))
	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "redmir",
	}))
}

func TestGroupRole(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}

	roles, err := m.ListAssignedRole(auth.Subject{
		Kind: auth.KindGroup,
		Name: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, roles, 1)
	assert.Equal(t, "admin", roles[0].Name)
}

func TestRoleGroup(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}

	user, err := m.CreateUser("bluemir", auth.WithGroup("admin"))
	if err != nil {
		t.Fatal(err)
	}

	role, err := m.GetRole("admin")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "admin", role.Name)

	result := role.IsAllow(auth.Context{
		Verb: "update",
		Subject: auth.Subject{
			Kind: auth.KindGroup,
			Name: "admin",
		},
		User: user,
		Resource: auth.Resource{
			"kind": "user",
			"name": "redmir",
		},
	})
	assert.Equal(t, true, result)
}

func TestIsSubsetOf(t *testing.T) {
	selector := auth.Resource{
		"kind": "user",
	}
	resource := auth.Resource{
		"kind": "user",
		"name": "redmir",
	}

	assert.True(t, selector.IsSubsetOf(resource))
}

func TestRule(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}
	rule := auth.Rule{
		//Verbs: []auth.Verb{"update"},
		Resource: auth.Resource{
			"kind": "user",
		},
	}

	user, err := m.CreateUser("bluemir", auth.WithGroup("admin"))
	if err != nil {
		t.Fatal(err)
	}

	result := rule.IsMatched(auth.Context{
		Verb: "update",
		Subject: auth.Subject{
			Kind: auth.KindGroup,
			Name: "admin",
		},
		User: user,
		Resource: auth.Resource{
			"kind": "user",
			"name": "redmir",
		},
	})
	assert.True(t, result)
}
