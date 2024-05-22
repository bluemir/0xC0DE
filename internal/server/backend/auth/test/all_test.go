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

	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "bluemir",
	}))
	assert.Equal(t, true, m.Can(user, "update", auth.Resource{
		"kind": "user",
		"name": "redmir",
	}))
}
