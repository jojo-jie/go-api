package server

import (
	"context"
	"encoding/json"
	"tag-service/pkg/bapi"
	pb "tag-service/proto"
)

type TagServer struct {
	pb.UnimplementedTagServiceServer
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (s *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	api:=bapi.NewAPI("http://127.0.0.1:4433")
	body, err := api.GetTagList(ctx, r.GetName())
	if err != nil {
		return nil, err
	}
	tagList := pb.GetTagListReply{}
	err=json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, err
	}
	return &tagList, nil
}
