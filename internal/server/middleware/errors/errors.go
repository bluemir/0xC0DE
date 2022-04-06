package errors

import (
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type HTTPErrorResponse struct {
	Message string   `json:"message"`
	Cause   []string `json:"cause,omitempty"`
}

func (e HTTPErrorResponse) String() string {
	if len(e.Cause) > 0 {
		return e.Message + "\n" + strings.Join(e.Cause, "\n")
	}
	return e.Message
}

type handlerOpts struct {
	showStackTrace bool
}
type Option func(*handlerOpts)

func ShowStackTrace(o *handlerOpts) {
	o.showStackTrace = true
}

func Handler(options ...Option) gin.HandlerFunc {
	opts := &handlerOpts{}
	for _, f := range options {
		f(opts)
	}

	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return // skip. no error
		}

		switch findRequestType(c) {
		case typeAPI:
			code, res := getResponse(c, opts)
			c.JSON(code, res)
			return
		case typeHTML:
			code, res := getResponse(c, opts)
			c.HTML(code, getHTMLName(code), res)
			return
		case typeUnknown:
			code, res := getResponse(c, opts)
			c.String(code, "text/plain", res.String())
		}
	}
}
func getResponse(c *gin.Context, opts *handlerOpts) (int, HTTPErrorResponse) {
	code := c.Writer.Status()
	if code < 100 {
		// find code..
		if c.Errors.Last().IsType(gin.ErrorTypeBind) {
			code = http.StatusBadRequest
		} else {
			code = findCode(c.Errors.Last())
		}
	}

	res := HTTPErrorResponse{}
	res.Message = c.Errors.Last().Err.Error()

	if !opts.showStackTrace {
		res.Cause = getStackTrace(c.Errors.Last())
	}
	return code, res
}

type reqType int

const (
	typeUnknown = iota
	typeAPI
	typeHTML
)

func findRequestType(c *gin.Context) reqType {
	for _, ct := range c.Accepted {
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil {
			continue
		}
		switch mt {
		case "application/json":
			return typeAPI
		case "text/html":
			return typeHTML
		}
	}
	return typeUnknown
}
func getHTMLName(code int) string {
	switch code {
	case http.StatusUnauthorized:
		return "/errors/unauthorized.html"
	case http.StatusNotFound:
		return "/errors/not-found.html"
	default:
		return "/errors/internal-sever-error.html"
	}
}
func getStackTrace(err error) []string {
	type StackTracer interface {
		StackTrace() errors.StackTrace
	}
	for {
		if st, ok := err.(StackTracer); ok {
			result := []string{}
			for _, f := range st.StackTrace() {
				// https://github.com/pkg/errors/blob/master/stack.go#L54-L57
				// https://github.com/pkg/errors/blob/master/stack.go#L86-L94
				buf, _ := f.MarshalText()
				result = append(result, string(buf))
			}
			return result
		}

		err = errors.Unwrap(err)
		if err == nil {
			return nil
		}
	}
}
