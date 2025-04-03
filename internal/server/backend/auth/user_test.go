package auth_test

import (
	"testing"

	"github.com/bluemir/0xC0DE/internal/functional"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/stretchr/testify/assert"
)

func TestCreateWithGroup(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.CreateUser("bluemir", auth.WithGroup("test-1")); err != nil {
		t.Fatal(err)
	}

	u, err := m.GetUser("bluemir")
	if err != nil {
		t.Fatal(err)
	}

	g := functional.First(u.Groups, func(g auth.Group) bool { return g.Name == "test-1" })
	assert.NotNil(t, g)
	assert.Equal(t, "test-1", g.Name)
}
