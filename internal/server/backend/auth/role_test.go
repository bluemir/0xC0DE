package auth_test

import (
	"testing"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/stretchr/testify/assert"
)

func TestRule(t *testing.T) {
	rule := auth.Rule{
		Verbs: []auth.Verb{"test"},
		Selector: auth.KeyValues{
			"foo": "bar",
		},
	}

	assert.True(t, rule.IsMatched(auth.Context{
		Verb: "test",
		Resource: auth.KeyValues{
			"foo": "bar",
		},
	}))
}
func TestEmptyRule(t *testing.T) {
	rule := auth.Rule{}

	assert.True(t, rule.IsMatched(auth.Context{
		Verb: "test",
		Resource: auth.KeyValues{
			"foo": "bar",
		},
	}))

}
