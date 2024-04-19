package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	pb "jgrpc/demo"
	"net"
)

var _ pb.GreeterServiceServer = &GreeterServiceServerImpl{}

type GreeterServiceServerImpl struct {
	pb.UnimplementedGreeterServiceServer
}

func (s *GreeterServiceServerImpl) Greet(ctx context.Context, request *pb.GreetRequest) (*pb.GreetResponse, error) {
	//TODO implement me
	greet, _ := json.Marshal(request.List)
	return &pb.GreetResponse{
		Greet: string(greet),
	}, nil
}

func main() {
	s := grpc.NewServer()
	pb.RegisterGreeterServiceServer(s, &GreeterServiceServerImpl{})
	listen, err := net.Listen("tcp", ":8009")
	if err != nil {
		panic(err)
	}
	err = s.Serve(listen)
	if err != nil {
		panic(err)
	}
}
