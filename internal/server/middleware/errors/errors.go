package errors

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Middleware(c *gin.Context) {
	c.Next()

	errs := c.Errors.ByType(gin.ErrorTypeAny)
	if len(errs) == 0 {
		return
	}

	if c.Writer.Written() && c.Writer.Size() > 0 {
		logrus.Tracef("response already written: %s", c.Errors.String())
		return // skip. already written
	}

	// Last one is most important
	err := c.Errors.Last()
	logrus.Errorf("%#v", err)

	code := code(err)
	logrus.Error(err, code)

	// with header or without header, or other processer/ maybe hook? depend on error type? or just code
	for _, accept := range strings.Split(c.Request.Header.Get("Accept"), ",") {
		t, _, e := mime.ParseMediaType(accept)
		if e != nil {
			logrus.Error(e)
			continue
		}

		switch t {
		case "application/json":
			// TODO make response json
			c.JSON(code, gin.H{
				"errors": c.Errors,
			})
			return
		case "text/html", "*/*":
			/* basic auth
			if code == http.StatusUnauthorized {
				c.Header(auth.LoginHeader(c.Request))
			}
			*/
			logrus.Trace(htmlName(err))
			c.HTML(code, htmlName(err), c.Errors)
			return
		case "text/plain":
			c.String(code, "%#v", c.Errors)
			return
		}
	}
	c.String(code, "%#v", c.Errors)
}
func code(err *gin.Error) int {
	if e := (sqlite3.Error{}); errors.As(err, &e) == true {
		switch e.Code {
		case sqlite3.ErrConstraint:
			return http.StatusBadRequest
		}
	}

	switch {
	case errors.Is(err, validator.ValidationErrors{}):
		return http.StatusBadRequest
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return http.StatusConflict
	case errors.Is(err, auth.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, os.ErrNotExist):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
func htmlName(err *gin.Error) string {
	switch {
	case errors.Is(err, validator.ValidationErrors{}):
		return "errors/bad-request.html"
	case errors.Is(err, auth.ErrUnauthorized):
		return "errors/unauthorized.html"
	case errors.Is(err, auth.ErrForbidden):
		return "errors/forbidden.html"
	case errors.Is(err, os.ErrNotExist):
		return "errors/not-found.html"
	}
	return "errors/internal-server-error.html"
}
