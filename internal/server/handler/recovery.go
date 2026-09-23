package handler

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

const SessionKeyRecovery = "recovery"

// RecoverySession 은 복구 링크를 통과한 상태다.
// 로그인한 것이 아니라 인증 수단을 다시 세울 수 있을 뿐이다.
type RecoverySession struct {
	Username string
}

func init() {
	gob.Register(&RecoverySession{})
}

// RequestRecovery 는 복구 메일을 보낸다.
//
// 계정이 있는지, 최근에 이미 보냈는지를 응답으로 드러내지 않는다.
// 무엇을 넣어도 같은 답을 주고, 발송은 백그라운드로 넘긴다.
// 여기서 SMTP 와 bcrypt 를 기다리면 계정이 있는 요청만 느려져 존재 여부가 샌다.
//
// @Router /api/v1/recover [post]
func RequestRecovery(c *gin.Context) error {
	req := struct {
		Email string `form:"email" json:"email" binding:"required,email"`
	}{}

	if err := c.ShouldBind(&req); err != nil {
		return err
	}

	if user, err := backends(c).Auth.FindUserByEmail(req.Email); err == nil {
		sendRecoveryMail(c, user.Name, user.Email)
	} else {
		logrus.Debugf("recovery requested for unknown email: %s", req.Email)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "등록된 주소라면 복구 메일을 보냈습니다. 메일함을 확인하세요.",
	})
	return nil
}

func sendRecoveryMail(c *gin.Context, username string, email string) {
	site := location.Get(c)
	manager := backends(c).Auth
	sender := mailer(c)

	if _, err := backends(c).Jobs.Run(c, "recovery-mail", func(ctx context.Context) error {
		secret, err := manager.IssueRecovery(username)
		if err != nil {
			return err
		}

		title, content, err := renderMail("recovery", recoveryMailData{
			AppName:  buildinfo.AppName,
			Username: username,
			Link:     recoveryLink(site, username, secret),
			Lifetime: auth.RecoveryLifetime,
		})
		if err != nil {
			return err
		}

		return sender.Send(ctx, sender.From(), []string{email}, title, content)
	}); err != nil {
		logrus.Errorf("fail to start recovery mail job: %+v", err)
	}
}

// recoveryLink 는 메일에 담을 복구 링크를 만든다.
// access key 와 같은 모양이다. secret 은 base64url 이라 마지막 점에서 갈라진다.
func recoveryLink(site *url.URL, username string, secret string) string {
	return fmt.Sprintf("%s://%s/users/recover/%s.%s",
		site.Scheme, site.Host, url.PathEscape(username), secret)
}

func splitRecoveryToken(token string) (string, string) {
	i := strings.LastIndex(token, ".")
	if i < 0 {
		return token, ""
	}
	return token[:i], token[i+1:]
}

type recoveryMailData struct {
	AppName  string
	Username string
	Link     string
	Lifetime time.Duration
}

// EnterRecovery 는 복구 링크를 검증하고 복구 세션을 연다.
// 링크를 여는 것만으로는 아무것도 바뀌지 않는다. 복구 비밀은 수단을 세운 뒤에 폐기한다.
func EnterRecovery(c *gin.Context) {
	username, secret := splitRecoveryToken(c.Param("token"))

	user, err := backends(c).Auth.ValidateRecovery(username, secret)
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	session := sessions.Default(c)
	session.Set(SessionKeyRecovery, &RecoverySession{Username: user.Name})
	if err := session.Save(); err != nil {
		c.Error(err)
		c.Abort()
		return
	}
}

// ResetPassword 는 복구 세션에서 비밀번호를 다시 설정한다.
//
// @Router /api/v1/recover/password [post]
func ResetPassword(c *gin.Context) error {
	recovery, ok := recoverySession(c)
	if !ok {
		return auth.ErrUnauthorized
	}

	req := struct {
		Password string `form:"password" json:"password" binding:"required,min=4"`
	}{}
	if err := c.ShouldBind(&req); err != nil {
		return err
	}

	if err := backends(c).Auth.UpdatePassword(recovery.Username, req.Password); err != nil {
		return err
	}

	user, err := finishRecovery(c, recovery.Username)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, user)
	return nil
}

// RecoveryStatus 는 복구 세션이 어느 계정의 것인지 알려준다. 화면이 이름을 보여주는 데 쓴다.
//
// @Router /api/v1/recover/session [get]
func RecoveryStatus(c *gin.Context) error {
	recovery, ok := recoverySession(c)
	if !ok {
		return auth.ErrUnauthorized
	}

	c.JSON(http.StatusOK, gin.H{"username": recovery.Username})
	return nil
}

func recoverySession(c *gin.Context) (*RecoverySession, bool) {
	recovery, ok := sessions.Default(c).Get(SessionKeyRecovery).(*RecoverySession)
	return recovery, ok
}

// finishRecovery 는 복구 비밀과 복구 세션을 지우고 그 계정으로 로그인시킨다.
func finishRecovery(c *gin.Context, username string) (*auth.User, error) {
	if err := backends(c).Auth.RevokeRecovery(username); err != nil {
		return nil, err
	}

	user, err := backends(c).Auth.GetUser(username)
	if err != nil {
		return nil, err
	}

	session := sessions.Default(c)
	session.Delete(SessionKeyRecovery)
	session.Set(SessionKeyUser, user)
	if err := session.Save(); err != nil {
		return nil, err
	}

	return user, nil
}
