package handler

import (
	"encoding/gob"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

const (
	SessionKeyUser = "user"

	ContextKeyManager = "__auth_manager__"
	ContextKeyUser    = "__auth_user__"
)

func init() {
	gob.Register(&auth.User{})
	gob.Register(&auth.Token{})
}

type ResourceGetter func(c *gin.Context) auth.Resource

func RequireLogin(c *gin.Context) {
	user, err := me(c)
	if err != nil {
		c.Error(err)
		c.Abort()
		//c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	c.Set(ContextKeyUser, user)
}

func Can(verb auth.Verb, r ResourceGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := me(c)
		if err != nil {
			c.Error(err)
			c.Abort()
			//c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		resource := r(c)

		if !backends(c).Auth.Can(user, verb, resource) {
			c.Error(auth.ErrForbidden)
			c.Abort()
			//c.AbortWithError(http.StatusForbidden, auth.ErrForbidden)
			return
		}
	}
}

func Register(c *gin.Context) {
	req := struct {
		Username string `form:"username"     validate:"required,min=4"`
		Password string `form:"password"     validate:"required,min=4"`
	}{}

	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	u, _, err := backends(c).Auth.Register(req.Username, req.Password, auth.WithGroup("user"))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

// @Router /api/v1/login [post]
func Login(c *gin.Context) {
	req := struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}{}

	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := backends(c).Auth.Default(req.Username, req.Password)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	session := sessions.Default(c)
	session.Set(SessionKeyUser, user)
	session.Save()

	c.JSON(http.StatusOK, user)
}

// @Router /api/v1/logout [get]
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(SessionKeyUser)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func me(c *gin.Context) (*auth.User, error) {
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
	user, err := backends(c).Auth.HTTP(c.Request)
	if err != nil {
		return nil, auth.ErrUnauthorized
	}
	c.Set(ContextKeyUser, user)
	return user, nil
}

func Me(c *gin.Context) {
	u, err := me(c)

	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	users, err := backends(c).Auth.ListUser()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Router /api/v1/groups [get]
func ListGroups(c *gin.Context) {
	groups, err := backends(c).Auth.ListGroup()
	if err != nil {
		c.Error(err)
		return
	}

	res := []GroupWithBindingRoles{}
	for _, group := range groups {
		roles, err := backends(c).Auth.ListAssignedRole(auth.Subject{
			Kind: auth.KindGroup,
			Name: group.Name,
		})
		if err != nil {
			c.Error(err)
			return
		}

		roleNames := []string{}

		for _, role := range roles {
			roleNames = append(roleNames, role.Name)
		}

		res = append(res, GroupWithBindingRoles{
			Group: group,
			Roles: roleNames,
		})
	}

	c.JSON(http.StatusOK, res)
}

type GroupWithBindingRoles struct {
	auth.Group

	Roles []string `json:"roles,omitempty"`
}

// @Router /api/v1/roles [get]
func ListRoles(c *gin.Context) {
	roles, err := backends(c).Auth.ListRole()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, roles)
}

func CanAPI(c *gin.Context) {
	verb := c.Param("verb")
	kind := c.Param("resource")

	user, _ := me(c)

	// TODO Setup HTTP cache..

	resource := auth.KeyValues{"kind": kind}
	for k, v := range c.Request.URL.Query() {
		resource[k] = strings.Join(v, ",")
	}

	if backends(c).Auth.Can(user, auth.Verb(verb), resource) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusForbidden)
	}
}

type AuthzRequest struct {
	Verb     auth.Verb     `json:"verb"`
	Resource auth.Resource `json:"resource"`
}

type AuthzResponse struct {
	Verb     auth.Verb     `json:"verb"`
	Resource auth.Resource `json:"resource"`
	Allowed  bool
	//Cause string
}

// @Summary authz request
// @Param "" body array AuthzRequest "request"
// @Router /api/v1/can [get]
func CanBulkAPI(c *gin.Context) {
	req := []AuthzRequest{}

	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		return
	}

	user, _ := me(c)

	res := []AuthzResponse{}
	for _, item := range req {
		allow := backends(c).Auth.Can(user, item.Verb, item.Resource)

		res = append(res, AuthzResponse{
			Verb:     item.Verb,
			Resource: item.Resource,
			Allowed:  allow,
		})
	}

	c.JSON(http.StatusOK, res)
}
