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
func ListPost(c *gin.Context) {
	items, err := backends(c).Posts.List(c.Request.Context(), meta.Limit(20))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ListResponse[posts.Post]{
		Items: items,
	})
}

func StreamPost(c *gin.Context) {
	items, err := backends(c).Posts.List(c.Request.Context())
	if err != nil && err != gorm.ErrRecordNotFound {
		c.Error(err)
		return
	}
	for _, post := range items {
		c.SSEvent("post", post)
		c.Writer.Flush() //will write header.
	}

	ch := backends(c).Events.WatchEvent("posts/created", c.Request.Context().Done())

	gone := c.Stream(func(w io.Writer) bool {
		if evt, ok := <-ch; ok {
			c.SSEvent("post", evt.Detail)
			return true // continue
		}
		return false // disconnect
	})
	if gone {
		logrus.Debug("client gone")
	}
}
