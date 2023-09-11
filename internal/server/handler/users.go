package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	u, err := backends(c).Auth.CreateUser(req.Username)
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
