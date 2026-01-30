package cache_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/middleware/cache"
)

func TestCacheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		opts     []cache.OptionFn
		expected string
	}{
		{
			name:     "Default",
			opts:     []cache.OptionFn{},
			expected: "",
		},
		{
			name:     "MaxAge",
			opts:     []cache.OptionFn{cache.MaxAge(60 * time.Second)},
			expected: "max-age=60",
		},
		{
			name:     "NoStore",
			opts:     []cache.OptionFn{cache.Disable},
			expected: "no-store",
		},
		{
			name:     "NoCache(AlwaysCheck)",
			opts:     []cache.OptionFn{cache.AlwaysCheckBeforeUseCache},
			expected: "no-cache",
		},
		{
			name:     "Public(Shared)",
			opts:     []cache.OptionFn{cache.Shared},
			expected: "public",
		},
		{
			name:     "Private(Local)",
			opts:     []cache.OptionFn{cache.ForLocalCache},
			expected: "private",
		},
		{
			name: "StaticFiie",
			opts: []cache.OptionFn{cache.ForStaicFile},
			// max-age=86400, public, stale-while-revalidate=...
			// The order depends on implementation append order.
			// code: directives = append(directives, fmt.Sprintf("max-age=%d", opt.MaxAge)) if Undefined/Public/Private
			// ForStaticFile sets Public.
			// expected: "public, max-age=86400, stale-while-revalidate=2592000"
			expected: "public, max-age=86400, stale-while-revalidate=2592000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			handler := cache.Set(tc.opts...)
			handler(c)

			assert.Equal(t, tc.expected, w.Header().Get("Cache-Control"))
		})
	}
}

func TestETag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Case 1: No If-None-Match
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	cache.ETag()(c)

	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag)
	assert.Equal(t, http.StatusOK, w.Code)

	// Case 2: Match
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("If-None-Match", etag)

	cache.ETag()(c)

	assert.Equal(t, http.StatusNotModified, c.Writer.Status())
	assert.True(t, c.IsAborted())

	// Case 3: No Match
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("If-None-Match", "wrong-etag")

	cache.ETag()(c)

	assert.Equal(t, http.StatusOK, w.Code) // Should verify it didn't abort with 304, status defaults 200
	assert.False(t, c.IsAborted())
}
