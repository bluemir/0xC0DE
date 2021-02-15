package auth

import (
	"github.com/bluemir/0xC0DE/pkg/store"
	"github.com/pkg/errors"
)

type Manager struct {
	*store.Store
}
type User struct {
	*store.Metadata
}
type Token struct {
}

func (m *Manager) Default(name, unhashedKey string) (*Token, error) {
	return nil, errors.Errorf("not implements")
}
