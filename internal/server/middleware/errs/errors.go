package errs

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	// All other errors → RFC 9457 Problem Details format
	// (with content negotiation: Override Header > Accept Header)
	statusCode := code(err)
	logrus.WithField("code", statusCode).Tracef("%T", err.Err)
	if statusCode >= 500 {
		logrus.Warnf("Server Error. code: %d, %s", statusCode, err)
	}

	responseFormat := negotiateResponseFormat(c.Request.Header)
	// applies all registered adjusters to determine final response format.
	for _, fn := range formatAdjusters {
		responseFormat = fn(responseFormat, err)
	}

	switch responseFormat {
	case ContentTypeJSON:
		c.JSON(statusCode, toJSON(err))
	case ContentTypeProblemJSON:
		problem := toProblemDetail(statusCode, err, c.Request.URL.Path)
		data, _ := json.Marshal(problem)
		c.Data(statusCode, ContentTypeProblemJSON, data)
	case ContentTypeHTML:
		c.HTML(statusCode, htmlName(statusCode, err), c.Errors)
	case ContentTypeText:
		c.String(statusCode, "%#v", c.Errors)
	}
}

const (
	ContentTypeJSON        = "application/json"         // RFC 6749 OAuth error
	ContentTypeProblemJSON = "application/problem+json" // RFC 9457 Problem Details
	ContentTypeHTML        = "text/html"
	ContentTypeText        = "text/plain"
)

// HeaderResponseFormat is the override header for forcing a specific response format.
const HeaderResponseFormat = "X-Response-Format"

// negotiateResponseFormat determines response format from request headers:
// 1. Override header (X-Response-Format) - highest priority
// 2. Accept header - fallback
func negotiateResponseFormat(header http.Header) string {
	// Check override header first
	if override := header.Get(HeaderResponseFormat); override != "" {
		switch strings.ToLower(override) {
		case "json":
			return ContentTypeProblemJSON
		case "html":
			return ContentTypeHTML
		case "text":
			return ContentTypeText
		}
	}

	// Fall back to Accept header negotiation
	return negotiateContentType(header.Get("Accept"))
}

// negotiateContentType parses Accept header and returns the preferred content type.
func negotiateContentType(acceptHeader string) string {
	for _, accept := range strings.Split(acceptHeader, ",") {
		t, _, err := mime.ParseMediaType(accept)
		if err != nil {
			continue
		}

		switch t {
		case "application/json", "application/problem+json":
			return ContentTypeProblemJSON
		case "text/html", "*/*":
			return ContentTypeHTML
		case "text/plain":
			return ContentTypeText
		}
	}
	return ContentTypeText // default fallback
}
