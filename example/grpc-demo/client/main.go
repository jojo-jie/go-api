package main

import (
	"context"
	"google.golang.org/grpc"
	pb "grpc-demo/proto/helloword"
	"io"
	"log"
)

func main() {
	conn, _ := grpc.Dial(":8000", grpc.WithInsecure())
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	err := SayHello(client)
	if err != nil {
		log.Fatalf("SayHello err:%v", err)
	}
}

func SayHello(client pb.GreeterClient) error {
	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "eddycjy",
	})
	if err != nil {
		return err
	}
	log.Printf("client.SayHello resp: %s", resp.Message)
	return nil
}

func SayList(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, err := client.SayList(context.Background(), r)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("resp: %v", resp)
	}
	return nil
}
