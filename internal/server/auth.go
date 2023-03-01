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

var (
	ErrUnauthroized = auth.ErrUnauthroized
)

func (server *Server) initAuth() error {
	s1, err := gorm.New(server.db, server.conf.Salt)
	if err != nil {
		return err
	}
	s2, err := inmemory.New(server.conf.Salt)
	if err != nil {
		return err
	}

	store := composite.Store{
		AuthUserStore:        s1,
		AuthTokenStore:       s1,
		AuthGroupStore:       s1,
		AuthRoleStore:        s2,
		AuthRoleBindingStore: s2,
	}

	m, err := auth.New(store, server.conf.Salt)
	if err != nil {
		return err
	}
	server.auth = m

	for name, key := range server.conf.InitUser {
		logrus.Tracef("init user: %s %s", name, key)
		if _, _, err := server.auth.Register(name, key); err != nil {
			return err
		}
	}

	// for session store
	gob.Register(&auth.User{})

	return nil
}
