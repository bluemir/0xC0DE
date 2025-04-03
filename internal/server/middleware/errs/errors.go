package errs

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
				"error": err.Err.Error(),
			})
			return
		case "text/html", "*/*":
			/* basic auth
			if code == http.StatusUnauthorized {
				c.Header(auth.LoginHeader(c.Request))
			}
			*/
			logrus.Trace(htmlName(code, err))
			c.HTML(code, htmlName(code, err), c.Errors)
			return
		case "text/plain":
			c.String(code, "%#v", c.Errors)
			return
		}
	}
	c.String(code, "%#v", c.Errors)
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
