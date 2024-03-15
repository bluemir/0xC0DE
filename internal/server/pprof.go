package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (server *Server) RunPprofServer(ctx context.Context, bind string) func() error {
	timeout := 5 * time.Second

	return func() error {
		s := http.Server{
			Addr:              bind,
			Handler:           http.DefaultServeMux, // https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/net/http/pprof/pprof.go
			ReadHeaderTimeout: 1 * time.Minute,
			WriteTimeout:      3 * time.Minute,
		}

		errc := make(chan error)
		go func() {
			defer close(errc)

			logrus.Infof("Listening and serving HTTP on '%s'", bind)
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errc <- err
			}
		}()

		select {
		case <-ctx.Done():
			logrus.Warn("shutting down gracefully, press Ctrl+C again to force")
		case err := <-errc:
			logrus.Errorf("listen: %s\n", err)
		}

		// nCtx for shutdown timeout only
		nCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := s.Shutdown(nCtx); err != nil {
			return errors.Wrapf(err, "Server forced to shutdown: ")
		}

		return nil
	}
}
