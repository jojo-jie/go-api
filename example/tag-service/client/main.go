package main

import (
	"context"
	"encoding/json"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"tag-service/internal/middleware"
	"tag-service/pkg/errcode"
	pb "tag-service/proto"
)

func main() {
	ctx := context.Background()
	clientConn, err := GetClientConn(ctx, "localhost:6699", []grpc.DialOption{grpc.WithBlock()})
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer clientConn.Close()
	tagServiceClient := pb.NewTagServiceClient(clientConn)
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{
		Name: "Go",
	})
	if err != nil {
		sts := errcode.FromError(err)
		details := sts.Details()
		if len(details) > 1 {
			detail := details[0].(*pb.Error)
			// 客户端内部业务错误码
			log.Fatalf("tagServiceClient.GetTagList err:%v code:%d msg:%s", details, detail.Code, detail.Message)
		}
		if sts.Code() == codes.DeadlineExceeded {
			log.Fatalf("%s", "timeout")
		}
	}
	log.Printf("resp %v", resp)
	body, _ := json.Marshal(resp)
	log.Printf("resp %s", string(body))
}

func GetClientConn(ctx context.Context, target string, opt []grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append(opt, grpc.WithInsecure())
	opts = append(opts, grpc.WithChainUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(middleware.UnaryContextTimeout(), grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(2),
			grpc_retry.WithCodes(
				codes.Unknown,
				codes.Internal,
				codes.DeadlineExceeded,
			),
		)),
	))
	opts = append(opts, grpc.WithChainStreamInterceptor(
		grpc_middleware.ChainStreamClient(middleware.StreamContextTimeout()),
	))
	return grpc.DialContext(ctx, target, opts...)
}
