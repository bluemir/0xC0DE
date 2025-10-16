package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/bluemir/0xC0DE/internal/server/backend/posts"
)

func CreatePost(c *gin.Context) {
	req := struct {
		Message string `form:"message"`
	}{}
	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		return
	}

	post, err := backends(c).Posts.Create(c.Request.Context(), req.Message)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, post)
}
func FindPost(c *gin.Context) error {
	req := struct {
		Query struct {
			meta.ListOption
			posts.Query
		}
	}{}

	if err := c.ShouldBindQuery(req.Query); err != nil {
		return err
	}

	list, err := backends(c).Posts.FindWithOption(c.Request.Context(), req.Query.Query, &req.Query.ListOption)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, list)
	return nil
}

func StreamPost(c *gin.Context) {
	list, err := backends(c).Posts.List(c.Request.Context(), meta.Limit(-1))
	if err != nil && err != gorm.ErrRecordNotFound {
		c.Error(err)
		return
	}
	for _, post := range list.Items {
		c.SSEvent("post", post)
		c.Writer.Flush() //will write header.
	}

	ch := backends(c).Events.Watch(posts.EventPostCreated{}, c.Request.Context().Done())
	gone := c.Stream(func(w io.Writer) bool {
		if evt, ok := <-ch; ok {
			c.SSEvent("post", evt.Detail.(posts.EventPostCreated).Post)
			return true // continue
		}
		return false // disconnect
	})
	if gone {
		logrus.Debug("client gone")
	}
}
