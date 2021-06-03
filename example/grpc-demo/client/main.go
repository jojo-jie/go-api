package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-demo/proto/helloword"
	"io"
	"log"
	"time"
)

func main() {
	conn, _ := grpc.Dial(":8000", grpc.WithInsecure())
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	/*err := SayHello(client)
	if err != nil {
		log.Fatalf("SayHello err:%v", err)
	}

	err = SayList(client, &pb.HelloRequest{
		Name: "eddycjy",
	})
	if err != nil {
		log.Fatalf("SayList err:%v", err)
	}*/
	/*err := SayRecord(client, &pb.HelloRequest{
		Name: "zzzhhhh",
	})
	if err != nil {
		log.Fatalf("SayRecord err:%v", err)
	}*/
	err:=SayRoute(client, &pb.HelloRequest{
		Name: "jojo",
	})
	if err != nil {
		log.Fatalf("SayRecord err:%v", err)
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

func SayRecord(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, err := client.SayRecord(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < 6; i++ {
		err := stream.Send(r)
		time.Sleep(2 * time.Second)
		if err != nil {
			return err
		}
	}
	fmt.Printf("now %d\n", time.Now().UnixNano())
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	log.Printf("resp msg:%v", resp)
	return nil
}

func SayRoute(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, err := client.SayRoute(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < 6; i++ {
		err = stream.Send(r)
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
		log.Printf("resp err:%v", resp)
	}

	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return nil
}
