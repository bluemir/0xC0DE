package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func Register(c *gin.Context) {
	req := struct {
		Username string `form:"username"     validate:"required,min=4"`
		Password string `form:"password"     validate:"required,min=4"`
		Confirm  string `form:"confirm"      validate:"required,eqfield=Password"`
	}{}

	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	u, err := backends(c).Auth.CreateUser(req.Username, auth.WithGroup("user"))
	if err != nil {
		c.Error(err)
		return
	}

	if _, err := backends(c).Auth.IssueToken(req.Username, req.Password); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func ListUsers(c *gin.Context) {
	users, err := backends(c).Auth.ListUser()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, users)
}

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

func ListRoles(c *gin.Context) {
	roles, err := backends(c).Auth.ListRole()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, roles)
}

func Can(c *gin.Context) {
	verb := c.Param("verb")
	resource := c.Param("resource")

	user, _ := me(c)

	// TODO Setup HTTP cache..

	if backends(c).Auth.Can(user, auth.Verb(verb), auth.KeyValues{"kind": resource}) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusForbidden)
	}
}
