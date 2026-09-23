package handler

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

const (
	SessionKeyPasskeyRegistration = "passkey_registration"
	SessionKeyPasskeyLogin        = "passkey_login"
)

func init() {
	gob.Register(&auth.PasskeyRegistration{})
}

// PasskeyRegisterBegin 은 passkey 등록 ceremony 를 시작한다.
// 로그인한 사용자는 자기 계정에 passkey 를 추가하고,
// 복구 세션이면 복구 중인 계정에 추가하고,
// 둘 다 아니면 아직 없는 username 으로 새로 가입한다.
//
// @Router /api/v1/passkeys/register/begin [post]
func PasskeyRegisterBegin(c *gin.Context) error {
	username := ""

	if user, err := me(c); err == nil {
		username = user.Name
	} else if recovery, ok := recoverySession(c); ok {
		// 메일로 계정을 증명했으므로 기존 계정에 붙이는 것을 허용한다
		username = recovery.Username
	} else {
		// gin 기본 validator 는 binding 태그를 본다. 검증 실패는 400 으로 내려간다.
		req := struct {
			Username string `form:"username" json:"username" binding:"required,min=4"`
		}{}
		if err := c.ShouldBind(&req); err != nil {
			return err
		}

		// 이미 있는 계정에 passkey 를 붙이려면 그 계정으로 로그인해야 한다.
		if _, err := backends(c).Auth.GetUser(req.Username); err == nil {
			return auth.ErrUnauthorized
		}

		username = req.Username
	}

	creation, registration, err := backends(c).Auth.BeginPasskeyRegistration(location.Get(c), username)
	if err != nil {
		return err
	}

	session := sessions.Default(c)
	session.Set(SessionKeyPasskeyRegistration, registration)
	if err := session.Save(); err != nil {
		return err
	}

	c.JSON(http.StatusOK, creation)
	return nil
}

// PasskeyRegisterFinish 는 authenticator 응답을 검증해 passkey 를 저장한다.
// 본문은 WebAuthn 응답이라 label 은 query 로 받는다.
//
// @Router /api/v1/passkeys/register/finish [post]
func PasskeyRegisterFinish(c *gin.Context) error {
	session := sessions.Default(c)

	registration, ok := session.Get(SessionKeyPasskeyRegistration).(*auth.PasskeyRegistration)
	if !ok {
		return auth.ErrUnauthorized
	}

	session.Delete(SessionKeyPasskeyRegistration)
	if err := session.Save(); err != nil {
		return err
	}

	user, token, err := backends(c).Auth.FinishPasskeyRegistration(
		location.Get(c), registration, c.Query("label"), c.Request)
	if err != nil {
		return err
	}

	// 복구 중이었다면 여기서 복구 비밀을 폐기한다. 링크는 한 번만 쓰인다.
	if recovery, ok := recoverySession(c); ok && recovery.Username == user.Name {
		if _, err := finishRecovery(c, user.Name); err != nil {
			return err
		}

		c.JSON(http.StatusOK, toPasskeyResponse(token))
		return nil
	}

	// passkey 로 가입한 경우 바로 로그인 상태로 만든다.
	session.Set(SessionKeyUser, user)
	if err := session.Save(); err != nil {
		return err
	}

	c.JSON(http.StatusOK, toPasskeyResponse(token))
	return nil
}

// @Router /api/v1/passkeys/login/begin [post]
func PasskeyLoginBegin(c *gin.Context) error {
	assertion, sessionData, err := backends(c).Auth.BeginPasskeyLogin(location.Get(c))
	if err != nil {
		return err
	}

	session := sessions.Default(c)
	session.Set(SessionKeyPasskeyLogin, sessionData)
	if err := session.Save(); err != nil {
		return err
	}

	c.JSON(http.StatusOK, assertion)
	return nil
}

// @Router /api/v1/passkeys/login/finish [post]
func PasskeyLoginFinish(c *gin.Context) error {
	session := sessions.Default(c)

	sessionData, ok := session.Get(SessionKeyPasskeyLogin).([]byte)
	if !ok {
		return auth.ErrUnauthorized
	}

	session.Delete(SessionKeyPasskeyLogin)
	if err := session.Save(); err != nil {
		return err
	}

	user, err := backends(c).Auth.FinishPasskeyLogin(location.Get(c), sessionData, c.Request)
	if err != nil {
		return err
	}

	session.Set(SessionKeyUser, user)
	if err := session.Save(); err != nil {
		return err
	}

	c.JSON(http.StatusOK, user)
	return nil
}

// @Router /api/v1/passkeys [get]
func ListPasskeys(c *gin.Context) error {
	user, err := me(c)
	if err != nil {
		return err
	}

	tokens, err := backends(c).Auth.ListPasskey(user.Name)
	if err != nil {
		return err
	}

	items := []PasskeyResponse{}
	for _, token := range tokens {
		items = append(items, toPasskeyResponse(&token))
	}

	c.JSON(http.StatusOK, ListResponse[PasskeyResponse]{Items: items})
	return nil
}

// @Router /api/v1/passkeys/{index} [delete]
func DeletePasskey(c *gin.Context) error {
	user, err := me(c)
	if err != nil {
		return err
	}

	req := struct {
		Index int `uri:"index"`
	}{}
	if err := c.ShouldBindUri(&req); err != nil {
		return err
	}

	if err := backends(c).Auth.RevokePasskey(user.Name, req.Index); err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
	return nil
}

// PasskeyResponse 는 Token 에서 공개해도 되는 값만 골라 담는다.
type PasskeyResponse struct {
	Index     int       `json:"index"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"createdAt"`
}

func toPasskeyResponse(token *auth.Token) PasskeyResponse {
	res := PasskeyResponse{
		Index:     token.Index,
		CreatedAt: token.CreatedAt,
	}
	if token.Secret.Passkey != nil {
		res.Label = token.Secret.Passkey.Label
	}
	return res
}
