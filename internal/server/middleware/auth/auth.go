package auth

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/bluemir/0xC0DE/internal/auth"
)

type Resource = auth.Resource
type Verb = auth.Verb
type KeyValues = auth.KeyValues

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
func User(c *gin.Context) (*auth.User, error) {
	// 1. try to get user from context
	if u, ok := c.Get(ContextKeyUser); ok {
		if user, ok := u.(*auth.User); ok {
			return user, nil
		}
	}

	// 2. next check session
	session := sessions.Default(c)
	u := session.Get(SessionKeyUser)
	if u != nil {
		if user, ok := u.(*auth.User); ok {
			c.Set(ContextKeyUser, user)
			return user, nil
		}
	}

	// 3. check api token or basic auth
	user, err := manager(c).HTTP(c.Request)
	if err != nil {
		return nil, auth.ErrUnauthorized
	}
	c.Set(ContextKeyUser, user)
	return user, nil
}

type ResourceGetter func(c *gin.Context) auth.Resource

func Authz(r ResourceGetter, verb Verb) gin.HandlerFunc {
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
func IssueToken(c *gin.Context) {
	req := struct {
		Username string
		Password string
	}{}

	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := manager(c).Default(req.Username, req.Password)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	t := time.Now().Add(6 * time.Hour)

	key, err := manager(c).GenerateToken(user.Name, auth.ExpiredAt(t))
	token := auth.ToHTTPToken(user.Name, key)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expiredAt": t.Format(time.RFC3339),
	})
}
func RevokeToken(c *gin.Context) {
	username, unhashedKey, err := auth.ParseHTTPRequest(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if err := manager(c).RevokeToken(username, unhashedKey); err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
}
