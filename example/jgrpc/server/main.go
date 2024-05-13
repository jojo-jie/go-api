package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	pb "jgrpc/demo"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// Greet Header和Trailer区别
// 根本区别：发送的时机不同！
//
// ✨ headers会在下面三种场景下被发送
//
// SendHeader() 被调用时（包含grpc.SendHeader和stream.SendHeader)
// 第一个响应被发送时
// RPC结束时（包含成功或失败）
// ✨ trailer会在rpc返回的时候，即这个请求结束的时候被发送
//
// 差异在流式RPC（streaming RPC）中比较明显：
//
// 因为trailer是在服务端发送完请求之后才发送的，所以client获取trailer的时候需要在stream.CloseAndRecv或者stream.Recv 返回非nil错误 (包含 io.EOF)之后
//
// 如果stream.CloseAndRecv之前调用stream.Trailer()获取的是空
func (s *GreeterServiceServerImpl) Greet(ctx context.Context, request *pb.GreetRequest) (*pb.GreetResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("from md %+v\n", md.Get("version"))
	}

	header := metadata.Join(md, metadata.Pairs("header-key", "client-greet-v1"))
	grpc.SendHeader(ctx, header)
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
					return status.New(codes.InvalidArgument,
						err.Error()).Err()
				}
			}
		}
		if err != nil {
			return status.New(codes.InvalidArgument,
				err.Error()).Err()
		}
		destination := orders[orderId.GetValue()].Destination.GetValue()
		shipment, found := combinedShipmentMap[destination]
		if found {
			ord := orders[orderId.GetValue()]
			shipment.OrderList = append(shipment.OrderList, &ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orders[orderId.GetValue()].Destination.GetValue()), Status: "Processed!"}
			ord := orders[orderId.GetValue()]
			comShip.OrderList = append(shipment.OrderList, &ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrderList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrderList))
				if err := stream.Send(&comb); err != nil {
					st := status.New(codes.InvalidArgument,
						"Order does not exist. order id: ")
					details, err := st.WithDetails(&epb.BadRequest_FieldViolation{
						Field:       "ID",
						Description: fmt.Sprintf("Order ID received is not valid"),
					})
					if err == nil {
						return details.Err()
					}

					return st.Err()
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}

func (s *GreeterServiceServerImpl) GetOrder(ctx context.Context, value *wrapperspb.StringValue) (*pb.Order, error) {
	//TODO implement me
	return &pb.Order{}, nil
}

func main() {
	/*creds, err := credentials.NewServerTLSFromFile("./x509/server.crt", "./x509/server.key")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderStreamServerInterceptor),
		grpc.Creds(creds),
	)
	pb.RegisterGreeterServiceServer(s, &GreeterServiceServerImpl{})
	listen, err := net.Listen("tcp", ":8009")
	if err != nil {
		panic(err)
	}
	err = s.Serve(listen)
	if err != nil {
		panic(err)
	}*/

	grpcPort, gwPort := ":8009", ":8010"
	go func() {
		lis, err := net.Listen("tcp", grpcPort)
		if err != nil {
			panic(err)
		}

		s := grpc.NewServer()

		pb.RegisterGreeterServiceServer(s, &GreeterServiceServerImpl{})
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1"+grpcPort,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	gwmux := runtime.NewServeMux()
	err = pb.RegisterGreeterServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(gwPort, gwmux)

	/*gwmux := runtime.NewServeMux()

	err := pb.RegisterGreeterServiceHandlerServer(context.Background(), gwmux, &GreeterServiceServerImpl{})
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8010", gwmux)*/
}

func orderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Pre-processing logic
	s := time.Now()

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Postprocessing logic
	log.Printf("Method: %s, req: %s, resp: %s, latency: %s\n",
		info.FullMethod, req, m, time.Now().Sub(s))

	return m, err
}

func orderStreamServerInterceptor(srv interface{},
	ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	// Pre-processing logic
	s := time.Now()
	nss := newWrappedStream(ss)
	// Invoking the StreamHandler to complete the execution of RPC invocation
	err := handler(srv, ss)

	// PostProcessing logic
	log.Printf("Method: %s, req: %+v, resp: %+v, latency: %s\n",
		info.FullMethod, nss.Recv, nss.Send, time.Now().Sub(s))

	return err
}

type wrappedStream struct {
	Recv []interface{}
	Send []interface{}
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	w.Recv = append(w.Recv, m)

	return err
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)

	w.Send = append(w.Send, m)

	return err
}

func newWrappedStream(s grpc.ServerStream) *wrappedStream {
	return &wrappedStream{
		make([]interface{}, 0),
		make([]interface{}, 0),
		s,
	}
}
