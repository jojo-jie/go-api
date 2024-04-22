package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	pb "jgrpc/demo"
	"net"
	"strconv"
	"strings"
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

func (s *GreeterServiceServerImpl) SearchOrders(value *wrapperspb.StringValue, stream pb.GreeterService_SearchOrdersServer) error {
	orders := make([]pb.Order, 0)
	items := make([]string, 0)
	items = append(items, "google", "amazon", "bing")
	for i := range 5 {
		id := i + 1
		idStr := strconv.Itoa(id)
		orders = append(orders, pb.Order{
			Id:          idStr,
			Items:       items,
			Description: "demo" + idStr,
		})
	}

	for _, order := range orders {
		for _, item := range order.Items {
			if strings.Contains(item, value.Value) {
				err := stream.Send(&order)
				if err != nil {
					return fmt.Errorf("error send: %v", err)
				}
			}
		}
	}
	return nil
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
