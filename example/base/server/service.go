package main

import (
	"errors"
	"fmt"
	"sync"
)

type Service interface {
	ServiceName() string
}

var services sync.Map

func AddService(service Service) {
	services.Store(service.ServiceName(), service)
}

var ErrServiceNotFound = errors.New("service not found")

func GetService(name string) (Service, error) {
	service, ok := services.Load(name)
	if !ok {
		return nil, ErrServiceNotFound
	}
	return service.(Service), nil
}

type HelloService interface {
	Service
	SayHello(input *Input) (*Output, error)
}

type UserService interface {
	Service
	GetUser(req *GetUserReq) (*GetUserResp, error)
}

type helloService struct {
}

func (h helloService) ServiceName() string {
	return "hello"
}

func (h helloService) SayHello(input *Input) (*Output, error) {
	fmt.Printf("say hello from "+input.Name)
	return &Output{Msg: "Hello " + input.Name, Data: input.Age}, nil
}

type userService struct {
}

func (u userService) ServiceName() string {
	return "user"
}

func (u userService) GetUser(req *GetUserReq) (*GetUserResp, error) {
	return &GetUserResp{ID: req.ID, Name: fmt.Sprintf("mock:%d", req.ID)}, nil
}

type Input struct {
	Name string
	Age int
}

type Output struct {
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

type GetUserReq struct {
	ID uint32
}

type GetUserResp struct {
	ID   uint32
	Name string
}
