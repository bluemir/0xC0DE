package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/util/httputil"
)

func TestHttpClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"foo": "bar"})
			return
		}
		if r.URL.Path == "/text" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("hello world"))
			return
		}
		if r.URL.Path == "/timeout" {
			time.Sleep(200 * time.Millisecond)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := httputil.NewHttpClient(
		httputil.WithTimeout(100*time.Millisecond),
		httputil.WithMaxIdleConns(10),
		httputil.WithMaxIdleConnsPerHost(5),
		httputil.MaxConnsPerHost(5),
	)

	// JSON request/response
	t.Run("JSON", func(t *testing.T) {
		res, err := client.NewRequest(http.MethodGet, ts.URL+"/json")
		assert.NoError(t, err)

		var data map[string]string
		code, err := res.ShouldBind(&data, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "bar", data["foo"])
	})

	// Text request/response
	t.Run("Text", func(t *testing.T) {
		res, err := client.NewRequest(http.MethodGet, ts.URL+"/text")
		assert.NoError(t, err)

		var data string
		code, err := res.ShouldBind(&data, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "hello world", data)
	})

	// Timeout
	t.Run("Timeout", func(t *testing.T) {
		_, err := client.NewRequest(http.MethodGet, ts.URL+"/timeout")
		assert.Error(t, err) // Should timeout
	})
}

func TestHttpClient_NewRequest_Options(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check header
		if r.Header.Get("X-Custom") != "foo" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Check cookie
		c, err := r.Cookie("session")
		if err != nil || c.Value != "123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := httputil.NewHttpClient()
	res, err := client.NewRequest(http.MethodGet, ts.URL,
		httputil.WithHeader(http.Header{"X-Custom": []string{"foo"}}),
		httputil.WithCookie([]*http.Cookie{{Name: "session", Value: "123"}}),
	)
	assert.NoError(t, err)
	code, _ := res.Raw()
	assert.Equal(t, http.StatusOK, code)
}
