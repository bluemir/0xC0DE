package server

import (
	"net"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	v1 "github.com/bluemir/0xC0DE/pkg/api/v1"
)

func (server *Server) RunGRPCServer(ctx context.Context, bind string) func() error {
	return func() error {
		grpcServer := grpc.NewServer()

		// TODO register GRPC service
		v1.RegisterHelloServiceServer(grpcServer, &HelloServiceServer{Server: server})

		// Starting grpc Server
		lis, err := net.Listen("tcp", bind)
		if err != nil {
			logrus.Fatalf("failed to listen: %v", err)
		}

		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
		}()

		logrus.Infof("grpc server run on %s", bind)

		return grpcServer.Serve(lis)
	}
}

type HelloServiceServer struct {
	*Server
	v1.UnimplementedHelloServiceServer
}

func (server *Server) grpcGatewayHandler(ctx context.Context, grpcBind string) (gin.HandlerFunc, error) {
	mux := runtime.NewServeMux()

	registerFuncs := []func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error{
		// TODO register GRPC Gateway
		v1.RegisterHelloServiceHandlerFromEndpoint,
	}
	for _, rf := range registerFuncs {
		if err := rf(
			ctx,
			mux,
			grpcBind,
			[]grpc.DialOption{grpc.WithInsecure()},
		); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	wsmux := wsproxy.WebsocketProxy(mux)

	return gin.WrapF(wsmux.ServeHTTP), nil
}

// import "google.golang.org/grpc/codes"
/*
func (server *HelloServiceServer) SayHello(ctx context.Context, req *v1.HelloRequest) (*v1.HelloReply, error) {
	return &v1.HelloReply{Message: "hello " + req.GetName()}, nil
	return nil, grpc.Errorf(codes.Unimplemented, "Unimplemented")
}
*/
