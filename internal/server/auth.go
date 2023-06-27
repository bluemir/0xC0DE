package server

import (
	"encoding/gob"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/bluemir/0xC0DE/internal/auth/store/composite"
	"github.com/bluemir/0xC0DE/internal/auth/store/gorm"
	"github.com/bluemir/0xC0DE/internal/auth/store/inmemory"
)

const (
	SessionKeyUser = "token"
)
const (
	ContextKeyUser = "user"
)

func initAuth(db *gorm.DB, salt string, initUser map[string]string) (*auth.Manager, error) {
	s1, err := gorm.New(db, salt)
	if err != nil {
		return nil, err
	}
	s2, err := inmemory.New(salt)
	if err != nil {
		return nil, err
	}

	store := composite.Store{
		AuthUserStore:        s1,
		AuthTokenStore:       s1,
		AuthGroupStore:       s1,
		AuthRoleStore:        s2,
		AuthRoleBindingStore: s2,
	}

	m, err := auth.New(store, salt)
	if err != nil {
		return nil, err
	}

	for name, key := range initUser {
		logrus.Tracef("init user: %s %s", name, key)
		if _, _, err := m.Register(name, key); err != nil {
			return nil, err
		}
	}

	// for session store
	gob.Register(&auth.User{})

	return m, nil
}
