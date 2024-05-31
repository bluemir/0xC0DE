package test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func TestAddRoleWithCondition(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.CreateRole("hello", []auth.Rule{
		{
			Verbs: []auth.Verb{"create", "get", "list"},
		},
		{
			Verbs: []auth.Verb{"create", "get", "list"},
			Conditions: []auth.Condition{
				`subject.kind == "user"`,
				`subject.name == resource.name`,
			},
		},
	})
	if err != nil {
		logrus.Tracef("%#v", err)
		t.Fatal(err)
	}
	assert.Equal(t, "hello", r.Name)
	assert.Len(t, r.Rules, 2)
}
func TestRule(t *testing.T) {
	m, err := newManager()
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.CreateRole("test", []auth.Rule{
		{
			Resource: auth.KeyValues{
				"kind": "user",
			},
			Verbs: []auth.Verb{"get"},
			Conditions: []auth.Condition{
				`subject.kind == "user"`,
				`subject.name == resource.name`,
			},
		},
	})
	if err != nil {
		logrus.Tracef("%#v", err)
		t.Fatal(err)
	}
	assert.Equal(t, "test", r.Name)

	assert.Equal(t, true, r.IsAllow(auth.Context{
		Subject: auth.Subject{
			Kind: "user",
			Name: "bluemir",
		},
		Verb: "get",
		Resource: auth.KeyValues{
			"kind": "user",
			"name": "bluemir",
		},
	}))

	assert.Equal(t, false, r.IsAllow(auth.Context{
		Subject: auth.Subject{
			Kind: "user",
			Name: "bluemir",
		},
		Verb: "get",
		Resource: auth.KeyValues{
			"kind": "user",
			"name": "admin",
		},
	}))
}
