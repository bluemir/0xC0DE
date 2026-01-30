package errs

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Middleware(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	if c.Writer.Written() && c.Writer.Size() > 0 {
		logrus.Tracef("response already written: %s", c.Errors.String())
		return // skip. already written
	}

	// Last one is most important
	err := c.Errors.Last()
	logrus.Errorf("%#v", err.Err)

	code := code(err)
	logrus.WithField("code", code).Tracef("%T", err.Err)

	if c.Writer.Written() && code != c.Writer.Status() {
		logrus.Debugf("Response code already written, expected '%d', but it was '%d'", code, c.Writer.Status())
		code = c.Writer.Status() // overwrite code with responed code.
	}

	if code >= 500 {
		logrus.Warnf("Server Error. code: %d, %s", code, err)
	}

	switch negotiate(c) {
	case "application/json":
		c.JSON(code, ProblemDetails{
			Type:   "about:blank",
			Title:  http.StatusText(code),
			Status: code,
			Detail: err.Error(),
		})
	case "text/html":
		/* basic auth
		if code == http.StatusUnauthorized {
			c.Header(auth.LoginHeader(c.Request))
		}
		*/
		logrus.Trace(htmlName(code, err))
		c.HTML(code, htmlName(code, err), c.Errors)
	default:
		c.String(code, "%#v", c.Errors)
	}
}

func negotiate(c *gin.Context) string {
	// 1. Check Accept Header
	for _, accept := range strings.Split(c.Request.Header.Get("Accept"), ",") {
		t, _, err := mime.ParseMediaType(accept)
		if err != nil {
			continue
		}

		switch t {
		case "application/json":
			return "application/json"
		case "text/html", "*/*":
			return "text/html"
		}
	}
	// 2. Default
	return "text/plain"
}

// rfc 7807
type ProblemDetails struct {
	Type     string `json:"type"`               // URI reference that identifies the problem type
	Title    string `json:"title"`              // Short, human-readable summary of the problem type
	Status   int    `json:"status"`             // HTTP status code
	Detail   string `json:"detail,omitempty"`   // Human-readable explanation specific to this occurrence of the problem
	Instance string `json:"instance,omitempty"` // URI reference that identifies the specific occurrence of the problem
}

type HttpStatusCodeProvider interface {
	StatusCode() int
}

func code(err *gin.Error) int {
	// errors.Is check same value, but errors.As check only its type.
	if e, ok := err.Err.(HttpStatusCodeProvider); ok {
		return e.StatusCode()
	}

	switch {
	case errors.As(err, &validator.ValidationErrors{}):
		return http.StatusBadRequest
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return http.StatusConflict
	case errors.Is(err, auth.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, os.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, meta.ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.As(err, &sqlite3.Error{}):
		e := sqlite3.Error{}
		errors.As(err, &e)
		switch e.Code {
		case sqlite3.ErrConstraint:
			switch e.ExtendedCode {
			case sqlite3.ErrConstraintUnique:
				return http.StatusConflict
			}
			return http.StatusBadRequest
		}
		return http.StatusNotImplemented
	}

	// finally check string match
	logrus.Trace(err.Error())
	switch {
	case strings.HasPrefix(err.Error(), "html/template: ") && strings.HasSuffix(err.Error(), " is undefined"):
		//html/template: ".*" is undefined
		return http.StatusNotImplemented
	}

	return http.StatusInternalServerError
}

func htmlName(code int, err *gin.Error) string {
	switch {
	//override
	case errors.Is(err, validator.ValidationErrors{}):
		return "errors/bad-request.html"
	}

	switch code {
	case http.StatusBadRequest:
		return "errors/bad-request.html"
	case http.StatusUnauthorized:
		return "errors/unauthorized.html"
	case http.StatusForbidden:
		return "errors/forbidden.html"
	case http.StatusNotFound:
		return "errors/not-found.html"
	}

	return "errors/internal-server-error.html"
}
