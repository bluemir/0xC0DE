package auth

import (
	"encoding/gob"

	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

const (
	SessionKeyUser = "user"

	ContextKeyManager = "__auth_manager__"
	ContextKeyUser    = "__auth_user__"
)

func Middleware(m *auth.Manager) gin.HandlerFunc {
	// for session store
	gob.Register(&auth.User{})
	gob.Register(&auth.Token{})

	return func(c *gin.Context) {
		c.Set(ContextKeyManager, m)
	}
}
func manager(c *gin.Context) *auth.Manager {
	return c.MustGet(ContextKeyManager).(*auth.Manager)
}

func RequireLogin(c *gin.Context) {
	user, err := User(c)
	if err != nil {
		c.Error(err)
		c.Abort()
		//c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	c.Set(ContextKeyUser, user)
}

type ResourceGetter func(c *gin.Context) auth.Resource

func Can(verb auth.Verb, r ResourceGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := User(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			//c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		resource := r(c)

		if !manager(c).Can(user, verb, resource) {
			c.Error(auth.ErrForbidden)
			c.Abort()
			//c.AbortWithError(http.StatusForbidden, auth.ErrForbidden)
			return
		}
	}
}
