package graceful

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
)

type httpServerOption struct {
	shutdownTimeout time.Duration
}

type httpServerOptionFn func(*httpServerOption)

func WithShutdownTimeout(d time.Duration) httpServerOptionFn {
	return func(opt *httpServerOption) {
		opt.shutdownTimeout = d
	}
}

func Run(ctx context.Context, s *http.Server, opts ...httpServerOptionFn) error {
	opt := httpServerOption{
		shutdownTimeout: 5 * time.Second,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	s.BaseContext = func(l net.Listener) context.Context {
		return ctx
	}

	// setup graceful server
	// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-with-context/server.go

	/*s := http.Server{
		Addr:              bind,
		Handler:           handler,
		ReadHeaderTimeout: 1 * time.Minute,
		WriteTimeout:      3 * time.Minute,

		// TLS
		TLSConfig: ...
	}*/

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.WithStack(err)
	}

	if s.TLSConfig != nil {
		l = tls.NewListener(l, s.TLSConfig)
	}

	errc := make(chan error)
	go func() {
		defer close(errc)

		logrus.Infof("Listening and serving HTTP on '%s'", s.Addr)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
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
	nCtx, cancel := context.WithTimeout(context.Background(), opt.shutdownTimeout)
	defer cancel()

	if err := s.Shutdown(nCtx); err != nil {
		return errors.Wrapf(err, "Server forced to shutdown")
	}

	return nil
}
