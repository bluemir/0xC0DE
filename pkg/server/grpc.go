package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/bluemir/0xC0DE/pkg/gen/api/v1"
)

type grpcImpl struct {
	*Server
	*v1.UnimplementedHelloServiceServer
}

func (server *Server) initGrpcService() error {
	server.grpc = grpc.NewServer()
	// TODO register GRPC service
	v1.RegisterHelloServiceServer(server.grpc, &grpcImpl{Server: server})

	return nil
}
func (server *Server) serveGrpc(c *gin.Context) {
	if c.Request.ProtoMajor == 2 && strings.HasPrefix(c.GetHeader("Content-Type"), "application/grpc") {
		server.grpc.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}

/*
// import "google.golang.org/grpc/codes"
func (server *grpcImpl) SayHello(ctx context.Context, req *v1.HelloRequest) (*v1.HelloReply, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "Unimplemented")
}
*/
