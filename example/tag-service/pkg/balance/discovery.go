package balance

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"strconv"
	"strings"
	"sync"
	"tag-service/pkg/weight"
	"time"
)

const schema = "grpclb"

type ServiceDiscovery struct {
	cli         *clientv3.Client //etcd client
	cc          resolver.ClientConn
	serviceList sync.Map
	lock        sync.Mutex
	prefix      string // 监视前缀
	lbName      string
}

func (s *ServiceDiscovery) Name() string {
	panic("implement me")
}

// Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *ServiceDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	log.Println("Build")
	s.cc = cc
	s.prefix = "/" + target.Scheme + "/" + target.Endpoint + "/"
	//根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), s.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	go s.watcher()
	return s, nil
}

func (s *ServiceDiscovery) Scheme() string {
	return schema
}

func (s *ServiceDiscovery) ResolveNow(options resolver.ResolveNowOptions) {
	log.Println("ResolveNow")
}

func (s *ServiceDiscovery) Close() {
	log.Println("Close")
	s.cli.Close()
}

// NewServiceDiscovery 新建服务发现
func NewServiceDiscovery(endpoints []string, lbName string) (resolver.Builder, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &ServiceDiscovery{
		cli:    client,
		lbName: lbName,
	}, nil
}

// watcher 监听前缀
func (s *ServiceDiscovery) watcher() {
	rch := s.cli.Watch(context.Background(), s.prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", s.prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
	log.Printf("watching prefix:==%s now...", s.prefix)
}

// SetServiceList 设置新的服务地址
func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	//获取服务地址
	addr := resolver.Address{Addr: strings.TrimPrefix(key, s.prefix)}
	if s.lbName == weight.Name {
		//获取服务地址权重
		nodeWeight, err := strconv.Atoi(val)
		if err != nil {
			nodeWeight = 1
		}
		addr = weight.SetAddrInfo(addr, weight.AddrInfo{Weight: nodeWeight})
	}
	s.serviceList.Store(key, addr)
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println("put key :", key, "val :", addr, "all:", s.getServices())
}

func (s *ServiceDiscovery) DelServiceList(key string) {
	s.lock.Lock()
	defer s.lock.Lock()
	s.serviceList.Delete(key)
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println("del key:", key)
}

func (s *ServiceDiscovery) getServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serviceList.Range(func(key, value interface{}) bool {
		addrs = append(addrs, value.(resolver.Address))
		return true
	})
	return addrs
}
