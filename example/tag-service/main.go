package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	pb "tag-service/proto"
	"tag-service/server"
)

func main()  {
	s:=grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	//grpcurl -plaintext -d '{"name":"Go"}' localhost:6699 TagService.GetTagList
	reflection.Register(s)
	lis,err:=net.Listen("tcp", ":6699")
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("server.Serve err: %v", err)
	}
}