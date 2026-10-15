package handler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/bluemir/0xC0DE/internal/buildinfo"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

// sendVerificationMail 은 확인 비밀을 발급하고 메일 발송을 백그라운드로 넘긴다.
// 직전에 보낸 지 얼마 안 됐으면 auth.ErrVerificationTooSoon 이고 아무것도 보내지 않는다.
//
// 발급은 백그라운드로 넘기지 않는다. 보냈는지를 응답에 담아야 하기 때문이다.
// 복구 메일(sendRecoveryMail)은 계정 존재 여부가 응답 시간으로 새지 않게 bcrypt 까지
// 백그라운드로 밀지만, 이쪽은 로그인한 사용자가 자기 주소에 보내는 것이라 샐 것이 없다.
//
// SMTP 는 백그라운드다. 보내지 못해도 저장은 되돌리지 않는다. 주소는 이미 바뀌었고
// 확인만 안 된 상태이며, 사용자는 간격이 지난 뒤 다시 저장해 재발송할 수 있다.
func sendVerificationMail(c *gin.Context, username string, email string) error {
	secret, err := backends(c).Auth.IssueVerification(username, email)
	if err != nil {
		return err
	}

	title, content, err := renderMail("verify-email", verificationMailData{
		AppName:  buildinfo.AppName,
		Username: username,
		Link:     verificationLink(location.Get(c), username, secret),
		Lifetime: auth.VerificationLifetime,
	})
	if err != nil {
		return err
	}

	sender := mailer(c)

	if _, err := backends(c).Jobs.Run(c, "verification-mail", func(ctx context.Context) error {
		return sender.Send(ctx, sender.From(), []string{email}, title, content)
	}); err != nil {
		logrus.Errorf("fail to start verification mail job: %+v", err)
	}

	return nil
}

// verificationLink 는 메일에 담을 확인 링크를 만든다. 복구 링크와 같은 모양이다.
func verificationLink(site *url.URL, username string, secret string) string {
	return fmt.Sprintf("%s://%s/users/verify-email/%s.%s",
		site.Scheme, site.Host, url.PathEscape(username), secret)
}

type verificationMailData struct {
	AppName  string
	Username string
	Link     string
	Lifetime time.Duration
}

// ConfirmEmail 은 확인 링크를 검증하고 주소를 확인된 것으로 표시한다.
//
// 복구 링크와 달리 아무 세션도 열지 않는다. 메일함을 가졌다는 것은 그 주소가 닿는다는
// 증명일 뿐이고, 로그인 수단이 아니다. 다른 기기에서 열어도 되도록 로그인도 요구하지 않는다.
func ConfirmEmail(c *gin.Context) {
	username, secret := splitLinkToken(c.Param("token"))

	user, err := backends(c).Auth.ConfirmVerification(username, secret)
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	// 이 브라우저가 같은 계정으로 로그인해 있으면 세션에 든 옛 User 를 갈아 끼운다.
	// 그러지 않으면 설정 화면이 계속 "확인되지 않음" 을 보여준다.
	session := sessions.Default(c)
	if current, ok := session.Get(SessionKeyUser).(*auth.User); ok && current.Name == user.Name {
		session.Set(SessionKeyUser, user)
		if err := session.Save(); err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		c.Set(ContextKeyUser, user)
	}
}
