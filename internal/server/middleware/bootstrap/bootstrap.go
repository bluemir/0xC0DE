package bootstrap

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/util"
)

// Bootstrap
// register admin user for initialize or other purpose.

var (
	tokenMutex     sync.RWMutex
	bootstrapToken = ""
	expireTimer    *time.Timer
)

func IssueBootstrapToken(ctx *gin.Context) {
	// TODO if has one or more admin, reject bootstraping
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if expireTimer != nil {
		expireTimer.Stop()
	}

	bootstrapToken = util.RandomString(32)
	expireTimer = time.AfterFunc(5*time.Minute, func() {
		tokenMutex.Lock()
		defer tokenMutex.Unlock()
		bootstrapToken = ""
	})

	logrus.Warnf("Bootstrap Token Issued. Token: '%s'", bootstrapToken)
}

func CheckBootstrapToken(ctx *gin.Context) {
	req := struct {
		Token string `form:"token" json:"token"`
	}{}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.Error(err)
		ctx.Abort()
		return
	}

	tokenMutex.RLock()
	currentToken := bootstrapToken
	tokenMutex.RUnlock()

	if currentToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bootstrap token expired"})
		ctx.Abort()
		return
	}

	if req.Token != currentToken {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bootstrap token not matched"})
		ctx.Abort()
		return
	}

	// continue next handler
	ctx.Next()

	if ctx.Writer.Status() == http.StatusOK {
		tokenMutex.Lock()
		bootstrapToken = ""
		if expireTimer != nil {
			expireTimer.Stop()
		}
		tokenMutex.Unlock()
	}
}
