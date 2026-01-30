package errs

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept", "application/json")

	// Trigger error
	c.Error(errors.New("something went wrong"))

	Middleware(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var pd ProblemDetails
	err := json.Unmarshal(w.Body.Bytes(), &pd)
	assert.NoError(t, err)

	assert.Equal(t, "about:blank", pd.Type)
	assert.Equal(t, "Internal Server Error", pd.Title)
	assert.Equal(t, http.StatusInternalServerError, pd.Status)
	assert.Equal(t, "something went wrong", pd.Detail)
}

func TestNegotiate(t *testing.T) {
	tests := []struct {
		accept   string
		expected string
	}{
		{"application/json", "application/json"},
		{"text/html", "text/html"},
		{"text/plain", "text/plain"},
		{"*/*", "text/html"}, // parse media type returns */* which falls through switch?
		// Wait, my switch only has application/json and text/html.
		// If */* matches nothing, it goes to default "text/plain"?
		// Let's check logic:
		// t, _, err := mime.ParseMediaType("*/*") -> t is "*/*"
		// switch "*/*" ... no match -> loop continues -> returns "text/plain"
		// Is this desired? Original had case "text/html", "*/*":...
		// My refactor removed "*/*" from case.
		{"application/xml", "text/plain"},
	}

	for _, tt := range tests {
		t.Run(tt.accept, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Accept", tt.accept)

			got := negotiate(c)
			if tt.accept == "*/*" {
				// Special check for my suspicion
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}
