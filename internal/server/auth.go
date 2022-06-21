package server

import (
	"encoding/gob"

	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/auth"
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
	a, err := auth.New(server.db, server.conf.Salt)
	if err != nil {
		return err
	}
	server.auth = a

	for name, key := range server.conf.InitUser {
		logrus.Tracef("init user: %s %s", name, key)
		if _, err := server.auth.Register(name, key); err != nil {
			return err
		}
	}

	// for session store
	gob.Register(&auth.User{})

	return nil
}
