package middleware

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"runtime/debug"
	"tag-service/pkg/errcode"
	"time"
)

func AccessLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	requestLog := "access request log: method: %s, begin_time: %d, request: %v"
	beginTime := time.Now().Local().Unix()
	log.Printf(requestLog, info.FullMethod, beginTime, req)
	resp, err = handler(ctx, req)
	if err != nil {
		return nil, err
	}
	endTime := time.Now().Local().Unix()
	responseLog := "access response log: method: %s, begin_time: %d, end_time: %d,response: %v"
	log.Printf(responseLog, info.FullMethod, beginTime, endTime, req)
	return
}

func ErrorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		errLog := "error log: method: %s, code: %v, message: %v, details: %v"
		s := errcode.FromError(err)
		log.Printf(errLog, info.FullMethod, s.Code(), s.Message(), s.Details())
		return nil, err
	}
	return
}

func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			recoveryLog := "recovery log: method: %s, message: %v, stack: %s"
			log.Printf(recoveryLog, info.FullMethod, e,string(debug.Stack()[:]))
		}
	}()
	return handler(ctx, req)
}

func HelloInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// server 一元拦截器
	resp, err = handler(ctx, req)
	if err != nil {
		return nil, err
	}
	return
}

func WorldInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// server 一元拦截器
	resp, err = handler(ctx, req)
	if err != nil {
		return nil, err
	}
	return
}

//metadata
