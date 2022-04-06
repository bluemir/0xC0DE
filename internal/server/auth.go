package server

import (
	"encoding/gob"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/auth"
	"github.com/bluemir/0xC0DE/internal/server/handler"
)

const (
	SessionKeyUser = "token"
)
const (
	ContextKeyUser = "user"
)

var (
	ErrUnauthroized = auth.ErrUnauthroized
	APIErrorHandler = handler.APIErrorHandler
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
	gob.Register(&auth.Token{})

	return nil
}

func (server *Server) authMiddleware(c *gin.Context) {
	// 1. check session
	session := sessions.Default(c)
	user := session.Get(SessionKeyUser)
	if user != nil {
		c.Set(ContextKeyUser, user)
		c.Next()
		return
	}

	// 2. check api token
	user, err := server.auth.HTTP(c.Request)
	if err != nil {
		return
	}

	c.Set(ContextKeyUser, user)
	c.Next()
	return
}
func User(c *gin.Context) (*auth.User, error) {
	u, ok := c.Get(ContextKeyUser)
	if !ok {
		return nil, ErrUnauthroized
	}
	user, ok := u.(*auth.User)
	if !ok {
		return nil, ErrUnauthroized
	}
	return user, nil
}

func (server *Server) authHTML(c *gin.Context) {
	_, err := User(c)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "/errors/unauthorized.html", c)
		c.Abort()
		return
	}
	// TODO authz
}
func (server *Server) authAPI(c *gin.Context) {
	_, err := User(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
		return
	}
	// TODO authz
}

func (server *Server) apiLogin(c *gin.Context) {
	req := &struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}{}
	if err := c.ShouldBind(req); err != nil {
		APIErrorHandler(c, err)
		return
	}

	logrus.Tracef("%#v", req)

	token, err := server.auth.Default(req.Username, req.Password)
	if err != nil {
		APIErrorHandler(c, err)
		return
	}

	session := sessions.Default(c)
	session.Set(SessionKeyUser, token)
	if err := session.Save(); err != nil {
		APIErrorHandler(c, err)
	}

	c.JSON(http.StatusOK, token)
}
func (server *Server) apiLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(SessionKeyUser)
	if err := session.Save(); err != nil {
		APIErrorHandler(c, err)
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (server *Server) Token(c *gin.Context) (*auth.Token, error) {
	session := sessions.Default(c)
	t, ok := session.Get(SessionKeyUser).(*auth.Token)
	if !ok {
		return nil, ErrUnauthroized
	}
	return t, nil
}
