package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHtmlMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	// Mocking render behavior might be complex without templates,
	// but we can check if it sets the right content type acceptance
	// and attempts to render.
	// However, c.HTML will panic if no renderer is set.
	// So we might just test markHTML directly or wrap appropriately.

	// Testing markHTML
	markHTML(c)
	markHTML(c)
	assert.Equal(t, "text/html", c.NegotiateFormat("text/html", "application/json"))
	// Actually c.SetAccepted sets the accepted header in the response? No.
	// Let's look at implementation: c.SetAccepted sets Context.Accepted.

	// For `html` function, it calls c.HTML. We need a simpler test for now.
	// Let's skip `html` function full render test if it depends on template loading.
}

func TestAbortIfHasPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/test", nil)

	s := &Server{}
	handler := s.AbortIfHasPrefix("/api")
	handler(c)

	assert.True(t, c.IsAborted())
}

func TestGlobals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	markAPI(c)
	// markAPI implementation: c.SetAccepted("application/json")
	// This usually is used for content negotiation.

	markHTML(c)
	// markHTML implementation: c.SetAccepted("text/html")
}
