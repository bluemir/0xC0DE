package errs

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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

// toProblemDetail converts any error to ProblemDetail.
func toProblemDetail(statusCode int, err error, instance string) *ProblemDetail {
	switch {
	// case ...:
	default:
		return &ProblemDetail{
			Type:     "about:blank",
			Title:    http.StatusText(statusCode),
			Status:   statusCode,
			Detail:   err.Error(),
			Instance: instance,
		}
	}
}

// toJSON converts any error to a JSON response.
// Returns appropriate response struct based on error type.
func toJSON(err error) any {
	switch {
	// case errors.As(err, ...):
	default:
		return defaultJSONErrorResponse{
			Error: err.Error(),
		}
	}
}

// ProblemDetail represents an RFC 9457 Problem Details response.
// See: https://www.rfc-editor.org/rfc/rfc9457.html
type ProblemDetail struct {
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Status     int            `json:"status,omitempty"`
	Detail     string         `json:"detail,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Extensions map[string]any `json:"-"`
}

// MarshalJSON implements json.Marshaler to include extensions in the output.
func (p *ProblemDetail) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type":  p.Type,
		"title": p.Title,
	}

	if p.Status != 0 {
		m["status"] = p.Status
	}
	if p.Detail != "" {
		m["detail"] = p.Detail
	}
	if p.Instance != "" {
		m["instance"] = p.Instance
	}

	for k, v := range p.Extensions {
		m[k] = v
	}

	return json.Marshal(m)
}

// OAuthResponse represents an RFC 6749 OAuth 2.0 error response.
// See: https://www.rfc-editor.org/rfc/rfc6749#section-5.2
type OAuthResponse struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

type defaultJSONErrorResponse struct {
	Error string
}
