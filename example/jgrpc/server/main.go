package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	pb "jgrpc/demo"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	orderBatchSize = 3
)

type Orders map[string]pb.Order

var orders Orders

func init() {
	orders = make(map[string]pb.Order)
	items := make([]string, 0)
	items = append(items, "google", "amazon", "bing")
	for i := range 5 {
		id := i + 1
		idStr := strconv.Itoa(id)
		orders[idStr] = pb.Order{
			Id:          idStr,
			Items:       items,
			Description: "ABCDEFG",
		}
	}
}

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

func (s *GreeterServiceServerImpl) UpdateOrders(stream pb.GreeterService_UpdateOrdersServer) error {
	ordersStr := "Updated Order IDs :  "
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&wrapperspb.StringValue{Value: "Orders processed " + strings.TrimRight(ordersStr, ", ")})
		}
		orders[order.Id] = *order
		log.Println("Order ID ", order.Id, ": Updated")
		ordersStr += order.Id + ", "
	}
}

func (s *GreeterServiceServerImpl) ProcessOrders(stream pb.GreeterService_ProcessOrdersServer) error {
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		if err == io.EOF {
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
		}
		if err != nil {
			log.Println(err)
			return err
		}
		destination := orders[orderId.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]
		if found {
			ord := orders[orderId.GetValue()]
			shipment.OrderList = append(shipment.OrderList, &ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orders[orderId.GetValue()].Destination), Status: "Processed!"}
			ord := orders[orderId.GetValue()]
			comShip.OrderList = append(shipment.OrderList, &ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrderList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrderList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}

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
