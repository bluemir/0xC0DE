package server

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWHelper(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Success case
	wSuccess := w(func(c *gin.Context) error {
		c.Status(http.StatusOK)
		return nil
	})

	wFailure := w(func(c *gin.Context) error {
		return errors.New("handler error")
	})

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		wSuccess(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failure", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		wFailure(c)

		assert.True(t, c.IsAborted())
		assert.Len(t, c.Errors, 1)
	})
}

func TestBodyReaderTweak(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ReadsAndRestoresBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		bodyContent := "test body content"
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString(bodyContent))

		bodyReaderTweak(c)

		// Check if body is still readable
		readBody, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		assert.Equal(t, bodyContent, string(readBody))
	})
}
