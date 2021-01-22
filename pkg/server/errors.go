package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HTTPError struct {
	Message string `example:"error message"`
}
type ErrHandlerOpt func(code int, message string) (int, string)

func APIErrorHandler(c *gin.Context, err error, opts ...ErrHandlerOpt) {
	logrus.Warn(err)

	code, message := defaultCodeMessage(err)

	for _, f := range opts {
		code, message = f(code, message)
	}

	c.JSON(code, HTTPError{Message: message})
	c.Abort()
}
func defaultCodeMessage(err error) (int, string) {
	switch v := err.(type) {
	case nil:
		return http.StatusInternalServerError, "unknown error"
	default:
		switch v {
		case gorm.ErrRecordNotFound:
			return http.StatusNotFound, v.Error()
		case gorm.ErrRegistered:
			return http.StatusConflict, v.Error()
		default:
			return http.StatusInternalServerError, v.Error()
		}
	}
}
func withCode(code int) ErrHandlerOpt {
	return func(c int, message string) (int, string) {
		return code, message
	}
}
func withMessage(message string) ErrHandlerOpt {
	return func(code int, m string) (int, string) {
		return code, message
	}
}
func withAdditionalMessage(message string) ErrHandlerOpt {
	return func(code int, m string) (int, string) {
		return code, fmt.Sprintf("%s: %s", message, m)
	}
}
