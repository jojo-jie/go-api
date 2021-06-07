package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"log"
	pb "tag-service/proto"
)

func main() {
	ctx := context.Background()
	clientConn, err := GetClientConn(ctx, "localhost:6699", []grpc.DialOption{grpc.WithBlock()})
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer clientConn.Close()
	tagServiceClient:=pb.NewTagServiceClient(clientConn)
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{
		Name: "Go",
	})
	if err != nil {
		log.Fatalf("tagServiceClient.GetTagList err:%v", err)
	}
	log.Printf("resp %v", resp)
	body,_:=json.Marshal(resp)
	log.Printf("resp %s", string(body))
}

func GetClientConn(ctx context.Context, target string, opt []grpc.DialOption) (*grpc.ClientConn, error) {
	opt = append(opt, grpc.WithInsecure())
	return grpc.DialContext(ctx, target, opt...)
}
