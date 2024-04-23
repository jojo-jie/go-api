package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	pb "jgrpc/demo"
	"log"
	"strconv"
	"time"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:8009", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("%+v", err)
		}
	}(conn)
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

	stream, err := client.SearchOrders(ctx, &wrapperspb.StringValue{
		Value: "google",
	})
	if err != nil {
		panic(err)
	}
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break
		}
		log.Println("Search Result: ", order)
	}

	streamC, err := client.UpdateOrders(ctx)
	if err != nil {
		return
	}
	for i2 := range 3 {
		if err := streamC.Send(&pb.Order{
			Id:          strconv.Itoa(i2 + 1),
			Items:       []string{"A", "B"},
			Description: "A with B",
			Price:       0.11,
			Destination: "ABC",
		}); err != nil {
			panic(err)
		}
	}

	res, err := streamC.CloseAndRecv()
	if err != nil {
		panic(err)
	}
	log.Printf("Update Orders Res : %s", res)

	streamP, err := client.ProcessOrders(ctx)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := streamP.Send(&wrapperspb.StringValue{Value: "1"}); err != nil {
			panic(err)
		}

		if err := streamP.Send(&wrapperspb.StringValue{Value: "4"}); err != nil {
			panic(err)
		}

		if err := streamP.CloseSend(); err != nil {
			panic(err)
		}
	}()

	for {
		combinedShipment, err := streamP.Recv()
		if err == io.EOF {
			break
		}
		log.Println("Combined shipment : ", combinedShipment)
	}
}
