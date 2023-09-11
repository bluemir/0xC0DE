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

	user, err := m.CreateUser("bluemir")
	if err != nil {
		t.Fatal(err)
	}
	group, err := m.CreateGroup("user")
	if err != nil {
		t.Fatal(err)
	}

	user.Groups.Add(group.Name)

	if err := m.UpdateUser(user); err != nil {
		t.Fatal(err)
	}

	role, err := m.CreateRole("user", []auth.Rule{
		{
			Resource: auth.Resource{
				"kind": "user",
			},
			Verbs: []auth.Verb{"update"},
			Conditions: []auth.Condition{
				`user.name == resource.name`,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	t.Log(role)

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

	user, err := m.CreateUser("bluemir")
	if err != nil {
		t.Fatal(err)
	}
	group, err := m.CreateGroup("admin")
	if err != nil {
		t.Fatal(err)
	}

	user.Groups.Add(group.Name)

	if err := m.UpdateUser(user); err != nil {
		t.Fatal(err)
	}

	role, err := m.CreateRole("admin", []auth.Rule{
		{
			Resource: auth.Resource{
				"kind": "user",
			},
			Verbs: []auth.Verb{"update"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(role)

	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "bluemir",
	}))
	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "redmir",
	}))
}
