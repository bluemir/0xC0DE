package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
)

func (server *Server) RunPprofServer(ctx context.Context, bind string) func() error {
	return func() error {
		return http.ListenAndServe(bind, nil)
	}
}
