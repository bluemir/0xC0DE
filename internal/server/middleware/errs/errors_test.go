package errs_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/middleware/errs"
)

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setup        func(*gin.Context)
		expectedCode int
		expectedBody string
	}{
		{
			name: "NoError",
			setup: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Unauthorized",
			setup: func(c *gin.Context) {
				c.Error(auth.ErrUnauthorized)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Forbidden",
			setup: func(c *gin.Context) {
				c.Error(auth.ErrForbidden)
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "InternalServerError",
			setup: func(c *gin.Context) {
				c.Error(errors.New("some random error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "AbortWithError",
			setup: func(c *gin.Context) {
				c.AbortWithError(http.StatusBadRequest, errors.New("bad request"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r := gin.New()
			r.Use(errs.Middleware)
			r.GET("/", func(c *gin.Context) {
				tc.setup(c)
			})

			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Accept", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBody)
			}
		})
	}
}
