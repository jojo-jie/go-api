package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-demo/proto/helloword"
	"io"
	"log"
	"net"
	"time"
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

func (s *GreeterServer) SayRecord(stream pb.Greeter_SayRecordServer) error {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("nows %d\n", time.Now().UnixNano())
			return stream.SendAndClose(&pb.HelloReply{
				Message: "say.record",
			})
		}
		if err != nil {
			return err
		}
		log.Printf("resp:%v", resp)
	}

	return nil
}

func (s *GreeterServer) SayRoute(stream pb.Greeter_SayRouteServer) error {
	n := 0
	for {
		err := stream.Send(&pb.HelloReply{
			Message: "say.route",
		})
		if err != nil {
			return err
		}
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		n++
		log.Printf("resp :%v", resp)
	}
	return nil
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
	lis, _ := net.Listen("tcp", port)
	server.Serve(lis)
}
