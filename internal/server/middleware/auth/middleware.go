package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bluemir/0xC0DE/internal/auth"
)

const (
	SessionKeyUser = "token"

	ContextKeyManager = "__auth_manager__"
	ContextKeyUser    = "__auth_user__"
)

func Middleware(m *auth.Manager) gin.HandlerFunc {
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
		c.AbortWithError(http.StatusUnauthorized, err)
	}

	c.Set(ContextKeyUser, user)
}

type ResourceGetter func(c *gin.Context) auth.Resource

func Can(verb Verb, r ResourceGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := User(c)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		resource := r(c)

		if !manager(c).Can(user, verb, resource) {
			c.AbortWithError(http.StatusForbidden, errors.New("Forbiddend"))
			return
		}
	}
}
