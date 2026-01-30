package graceful_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/server/graceful"
)

func TestRun(t *testing.T) {
	// Use port 0 to let OS choose free port
	server := &http.Server{
		Addr: ":0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	errc := make(chan error)
	go func() {
		// Run should block until context cancellation
		errc <- graceful.Run(ctx, server, graceful.WithShutdownTimeout(1*time.Second))
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Since we don't know the port (random), we can't easily make a request *to* it
	// unless we parse logs or if Run exposed the listener (it doesn't return it).
	// But we can test the shutdown mechanism.

	// Trigger shutdown
	cancel()

	select {
	case err := <-errc:
		// Should return nil on graceful shutdown
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}

func TestRunWithError(t *testing.T) {
	// Test with invalid address to trigger immediate error
	server := &http.Server{
		Addr: "invalid-address",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := graceful.Run(ctx, server)
	assert.Error(t, err)
}
