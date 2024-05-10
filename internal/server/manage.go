package server

import (
	"context"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/bluemir/0xC0DE/internal/server/middleware/prom"

	// swagger
	_ "github.com/bluemir/0xC0DE/internal/swagger"
)

func (server *Server) RunManageServer(ctx context.Context, bind string) func() error {

	return func() error {
		// starting http server
		app := gin.New()

		// prometheus for monitoring
		app.GET("/metric", prom.Handler())

		// swagger
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

		return runGracefulServer(ctx, bind, app, nil)
	}
}
