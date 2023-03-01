package server

import (
	"encoding/gob"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/bluemir/0xC0DE/internal/auth/store/gorm"
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
	store, err := gorm.New(server.db, server.conf.Salt)
	if err != nil {
		return err
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
