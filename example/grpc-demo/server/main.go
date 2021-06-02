package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	pb "grpc-demo/proto/helloword"
	"net"
)

var port string

func init() {
	flag.StringVar(&port, "p", ":8000", "启动端口号")
	flag.Parse()
}

type GreeterServer struct {
	pb.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello world",
	}, nil
}

func (s *GreeterServer) SayList(r *pb.HelloRequest, stream pb.Greeter_SayListServer) error {
	for i := 0; i < 6; i++ {
		err := stream.Send(&pb.HelloReply{
			Message: "hello.list",
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	lis, _ := net.Listen("tcp", port)
	server.Serve(lis)
}
