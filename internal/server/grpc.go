package server

import (
	"net"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	v1 "github.com/bluemir/0xC0DE/pkg/api/v1"
)

func (server *Server) RunGRPCServer() error {
	grpcServer := grpc.NewServer()

	// TODO register GRPC service
	v1.RegisterHelloServiceServer(grpcServer, &HelloServiceServer{Server: server})

	// Starting grpc Server
	lis, err := net.Listen("tcp", server.conf.GRPCBind)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	return grpcServer.Serve(lis)
}

type HelloServiceServer struct {
	*Server
	v1.UnimplementedHelloServiceServer
}

func (server *Server) grpcGatewayMiddleware() (gin.HandlerFunc, error) {
	mux := runtime.NewServeMux()

	registerFuncs := []func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error{
		// TODO register GRPC Gateway
		v1.RegisterHelloServiceHandlerFromEndpoint,
	}
	//
	for _, rf := range registerFuncs {
		if err := rf(
			context.Background(),
			mux,
			server.conf.GRPCBind,
			[]grpc.DialOption{grpc.WithInsecure()},
		); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	wsmux := wsproxy.WebsocketProxy(mux)

	return func(c *gin.Context) {
		wsmux.ServeHTTP(c.Writer, c.Request)
		if c.Writer.Written() {
			c.Abort()
			return
		}
	}, nil
}

// import "google.golang.org/grpc/codes"
/*
func (server *HelloServiceServer) SayHello(ctx context.Context, req *v1.HelloRequest) (*v1.HelloReply, error) {
	return &v1.HelloReply{Message: "hello " + req.GetName()}, nil
	return nil, grpc.Errorf(codes.Unimplemented, "Unimplemented")
}
*/
