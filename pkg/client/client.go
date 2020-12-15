package client

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	v1 "github.com/bluemir/0xC0DE/pkg/gen/api/v1"
)

type Config struct {
	Endpoint string
}

func Run(conf *Config) error {
	conn, err := grpc.Dial(conf.Endpoint, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "connection failed")
	}
	defer conn.Close()

	client := v1.NewHelloServiceClient(conn)
	res, err := client.SayHello(context.Background(), &v1.HelloRequest{Name: "tom"})
	if err != nil {
		return errors.Wrap(err, "fail to call say hello")
	}
	logrus.Info(res.GetMessage())
	return nil
}
