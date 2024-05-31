package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/bluemir/0xC0DE/internal/server/graceful"
	"github.com/bluemir/0xC0DE/internal/server/handler"
	"github.com/bluemir/0xC0DE/internal/server/middleware/prom"

	// swagger
	_ "github.com/bluemir/0xC0DE/internal/swagger"
)

func (server *Server) RunAdminHTTPServer(ctx context.Context, bind string) func() error {
	return func() error {
		// starting http server
		app := gin.New()

		// ping
		app.GET("/ping", handler.Ping)

		// prometheus for monitoring
		app.GET("/metric", prom.Handler())

		// swagger
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

		// pprof
		// https://cs.opensource.google/go/go/+/refs/tags/go1.21.5:src/net/http/pprof/pprof.go
		app.Any("/debug/pprof/", gin.WrapF(pprof.Index))
		app.Any("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		app.Any("/debug/pprof/profile", gin.WrapF(pprof.Profile))
		app.Any("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
		app.Any("/debug/pprof/trace", gin.WrapF(pprof.Trace))

		return graceful.Run(ctx, &http.Server{
			Addr:    bind,
			Handler: app,
		})
	}
}
