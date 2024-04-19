package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "jgrpc/demo"
	"log"
	"time"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:8009", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewGreeterServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	list := make(map[string]string)
	list["id"] = "1"
	list["t"] = "cc"
	greet, err := client.Greet(ctx, &pb.GreetRequest{
		Name:     "jojo",
		Snippets: make([]string, 0),
		List:     list,
	})
	if err != nil {
		panic(err)
	}
	log.Print("Greet Response -> : ", greet)
}
