package main

import (
	"cmp"
	"context"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	pb "jgrpc/demo"
	"log"
	"os"
	"slices"
	"strconv"
	"time"
)

var header, trailer metadata.MD

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	creds, err := credentials.NewClientTLSFromFile("./x509/server.crt", "www.demo.com")
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial("127.0.0.1:8009",
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(orderStreamClientInterceptor),
		grpc.WithBlock(),
	)
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
	list := make(map[string]string)
	list["id"] = "1"
	list["t"] = "cc"
	mdBase := metadata.New(map[string]string{"version": "v1"})
	mdF := metadata.Pairs("version", "00001")
	md := metadata.Join(mdBase, mdF)
	log.Printf("matadata info %+v", md)
	ctx = metadata.NewOutgoingContext(ctx, md) //AppendToOutgoingContext
	greet, err := client.Greet(ctx, &pb.GreetRequest{
		Name:     "jojo",
		Snippets: make([]string, 0),
		List:     list,
	}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		panic(err)
	}
	log.Printf("Greet Response -> : %+v\n, header ->: %+v", greet, header)
	os.Exit(-1)

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
		st, ok := status.FromError(err)
		if !ok {
			log.Println(err)
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			for _, d := range st.Details() {
				switch info := d.(type) {
				case *epb.BadRequest_FieldViolation:
					log.Printf("Request Field Invalid: %s", info)
				default:
					log.Printf("Unexpected error type: %s", info)
				}
			}
		default:
			log.Printf("Unhandled error : %s ", st.String())
		}

		return
	}
	go func() {
		if err := streamP.Send(&wrapperspb.StringValue{Value: "1"}); err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.InvalidArgument {
				log.Println(st.Code(), st.Message())
			} else {
				log.Println(err)
			}
		}

		if err := streamP.Send(&wrapperspb.StringValue{Value: "4"}); err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.InvalidArgument {
				log.Println(st.Code(), st.Message())
			} else {
				log.Println(err)
			}
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

	orders := []Order{
		{"foo", "alice", 1.00},
		{"bar", "bob", 3.00},
		{"baz", "carol", 4.00},
		{"foo", "alice", 2.00},
		{"bar", "carol", 1.00},
		{"foo", "bob", 4.00},
	}

	slices.SortFunc(orders, func(a, b Order) int {
		return cmp.Or(
			cmp.Compare(a.Customer, b.Customer),
			cmp.Compare(a.Product, b.Product),
			cmp.Compare(b.Price, a.Price),
		)
	})
}

type Order struct {
	Product  string
	Customer string
	Price    float64
}

func orderUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Pre-processor phase
	s := time.Now()

	// Invoking the remote method
	err := invoker(ctx, method, req, reply, cc, opts...)

	// Post-processor phase
	log.Printf("method: %s, req: %s, resp: %s, latency: %s\n",
		method, req, reply, time.Now().Sub(s))

	return err
}

type wrappedStream struct {
	method string
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	log.Printf("method: %s, res: %s\n", w.method, m)

	return err
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)

	log.Printf("method: %s, req: %s\n", w.method, m)

	return err
}

func newWrappedStream(method string, s grpc.ClientStream) *wrappedStream {
	return &wrappedStream{
		method,
		s,
	}
}

func orderStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	// Pre-processing logic
	s := time.Now()

	cs, err := streamer(ctx, desc, cc, method, opts...)

	// PostProcessing logic
	log.Printf("method: %s, latency: %s\n", method, time.Now().Sub(s))

	return newWrappedStream(method, cs), err
}
