package main

import (
	"flag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	pb "tag-service/proto"
	"tag-service/server"
)

var grpcPort string
var httpPort string

func init() {
	flag.StringVar(&grpcPort, "grpc_port", "6699", "grpc启动端口号")
	flag.StringVar(&httpPort, "http_port", "0033", "http启动端口号")
	flag.Parse()
}

func main() {
	g := new(errgroup.Group)
	g.Go(func() error {
		return RunHttpServer(httpPort)
	})
	g.Go(func() error {
		return RunGrpcServer(grpcPort)
	})
	if err := g.Wait(); err != nil {
		log.Fatalf("Run server err:%v", err)
	}
}

func RunHttpServer(port string) error {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("pong"))
	})
	return http.ListenAndServe(":"+port, serveMux)
}

func RunGrpcServer(port string) error {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	//grpcurl -plaintext -d '{"name":"Go"}' localhost:6699 TagService.GetTagList
	reflection.Register(s)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
