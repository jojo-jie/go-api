package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/resolver"
	"tag-service/pkg/balance"
	/*"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/naming"*/
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"log"
	"tag-service/global"
	"tag-service/internal/middleware"
	"tag-service/pkg/errcode"
	"tag-service/pkg/tracer"
	pb "tag-service/proto"
	//"time"
)

type Auth struct {
	AppKey    string
	AppSecret string
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_key": a.AppKey, "app_secret": a.AppSecret}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func init() {
	err := setupTracer()
	if err != nil {
		log.Fatalf("article-service init.setupTracer err: %v\n", err)
	}
}

func main() {
	auth := Auth{
		AppKey:    "eddycjy",
		AppSecret: "go-programming-tour-book",
	}
	ctx := context.Background()
	/*md := metadata.New(map[string]string{"go": "programming", "tour": "book"})
	newCtx := metadata.NewOutgoingContext(ctx, md)*/
	newCtx := metadata.AppendToOutgoingContext(ctx, "go", "programming")
	clientConn, err := GetClientConn(newCtx, "tag-service", []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&auth)})
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer clientConn.Close()
	tagServiceClient := pb.NewTagServiceClient(clientConn)
	resp, err := tagServiceClient.GetTagList(newCtx, &pb.GetTagListRequest{
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

func GetClientConn(ctx context.Context, serviceName string, opt []grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append(opt, grpc.WithInsecure())
	opts = append(opts, grpc.WithChainUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(
			middleware.UnaryContextTimeout(),
			middleware.ClientTracing(),
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithMax(2),
				grpc_retry.WithCodes(
					codes.Unknown,
					codes.Internal,
					codes.DeadlineExceeded,
				),
			),
		),
	))
	opts = append(opts, grpc.WithChainStreamInterceptor(
		grpc_middleware.ChainStreamClient(middleware.StreamContextTimeout()),
	))
	discovery, err := balance.NewServiceDiscovery([]string{"http://localhost:2379"})
	if err != nil {
		return nil, err
	}
	resolver.Register(discovery)
	target := fmt.Sprintf("%s:///%s", discovery.Scheme(), serviceName)
	opts = append(opts, grpc.WithBlock(), grpc.WithBalancerName(grpc.PickFirstBalancerName))
	return grpc.DialContext(ctx, target, opts...)
}

func setupTracer() error {
	jaegerTracer, closer, err := tracer.NewJaegerTracer("article-service", "127.0.0.1:6831")
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	global.Closer = closer
	return nil
}
