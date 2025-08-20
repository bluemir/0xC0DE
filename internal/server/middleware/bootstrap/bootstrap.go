package bootstrap

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/util"
)

// Bootstrap
// register admin user for initialize or other purpose.

var (
	bootstrapToken = "" // QUESTION need lock?
)

func IssueBootstrapToken(c *gin.Context) {
	// TODO if has one or more admin, reject bootstraping
	bootstrapToken = util.RandomString(32)

	go time.AfterFunc(5*time.Minute, func() {
		bootstrapToken = ""
	})

	logrus.Warnf("Bootstrap Token Issued. Token: '%s'", bootstrapToken)
}
func CheckBootstrapToken(c *gin.Context) {
	req := struct {
		Token string `form:"token" json:"token"`
	}{}

	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	if bootstrapToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bootstrap token expired"})
		c.Abort()
		return
	}

	if req.Token != bootstrapToken {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bootstrap token not matched"})
		c.Abort()
		return
	}

	// continue next handler
	c.Next()

	if c.Writer.Status() == 200 {
		// reset token
		bootstrapToken = ""
	}
}
