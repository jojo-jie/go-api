package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	pb "jgrpc/demo"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
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
	select {
	case <-time.After(3 * time.Second):
		return nil, status.Error(codes.DeadlineExceeded, "timeout")
	case <-ctx.Done():
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
	return &pb.Order{
		Destination: value,
	}, nil
}

func (s *GreeterServiceServerImpl) AddOrder(ctx context.Context, order *pb.Order) (*anypb.Any, error) {
	log.Printf("%+v\n", order)
	return anypb.New(order)
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

func transform[A, B any](xs []A, f func(A) B) []B {
	ret := make([]B, len(xs))
	g := new(errgroup.Group)
	for i, x := range xs {
		g.Go(func() error {
			ret[i] = f(x)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		panic(err)
	}
	return ret
}

func atomicAs() {
	var ops atomic.Uint64
	var g errgroup.Group
	for i := 0; i < 50; i++ {
		g.Go(func() error {
			for c := 0; c < 1000; c++ {
				ops.Add(1)
			}
			return nil
		})
	}
	g.Wait()
	fmt.Println("ops:", ops.Load())

	data := "5oiR5Lus5LiN5piv54mb6ams5oiR5Lus5piv5Lq6"
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Println(sEnc)
	sDec, _ := base64.StdEncoding.DecodeString(sEnc)
	fmt.Println(string(sDec))
	fmt.Println()
	uEnc := base64.URLEncoding.EncodeToString([]byte(data))
	fmt.Println(uEnc)
	uDec, _ := base64.URLEncoding.DecodeString(uEnc)
	fmt.Println(string(uDec))

	jobs := make(chan int, 5)
	done := make(chan bool)
	go func() {
		for {
			j, more := <-jobs
			if more {
				fmt.Println("received job", j)
			} else {
				fmt.Println("received all jobs")
				done <- true
				return
			}
		}
	}()
	for j := 1; j <= 3; j++ {
		jobs <- j
		fmt.Println("sent job", j)
	}
	close(jobs)
	fmt.Println("sent all jobs")
	<-done
	_, ok := <-jobs
	fmt.Println("received more jobs:", ok)

	nextInt := intSeq()
	fmt.Println(nextInt())

	newInts := intSeq()
	fmt.Println(newInts())
}

func intSeq() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

type List[T any] struct {
	head, tail *element[T]
}

type element[T any] struct {
	next *element[T]
	val  T
}

func (lst *List[T]) Push(v T) {
	if lst.tail == nil {
		lst.head = &element[T]{val: v}
		lst.tail = lst.head
	} else {
		lst.tail.next = &element[T]{val: v}
		lst.tail = lst.tail.next
	}
}

func (lst *List[T]) GetAll() []T {
	var elems []T
	for e := lst.head; e != nil; e = e.next {
		elems = append(elems, e.val)
	}
	return elems
}

type Service interface {
	HelloWorld(name string) (string, error)
}

type service struct{}

func (s service) HelloWorld(name string) (string, error) {
	return fmt.Sprintf("Hello World from %s", name), nil
}

type validator struct {
	next Service
}

func (v validator) HelloWorld(name string) (string, error) {
	if len(name) <= 3 {
		return "", fmt.Errorf("name length must be greater than 3")
	}

	return v.next.HelloWorld(name)
}

type logger struct {
	next Service
}

func (l logger) HelloWorld(name string) (string, error) {
	res, err := l.next.HelloWorld(name)

	if err != nil {
		fmt.Println("error:", err)
		return res, err
	}

	fmt.Println("HelloWorld method executed successfully")
	return res, err
}

func New() Service {
	return logger{
		next: validator{
			next: service{},
		},
	}
}
