package balance

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
	"log"
	"sync"
	"time"
)

const schema = "grpclb"

type ServiceDiscovery struct {
	cli         *clientv3.Client //etcd client
	cc          resolver.ClientConn
	serviceList map[string]resolver.Address
	lock        sync.Mutex
}

// Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *ServiceDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	log.Println("Build")
	s.cc = cc
	s.serviceList = make(map[string]resolver.Address)
	prefix := "/" + target.Scheme + "/" + target.Endpoint + "/"
	//根据前缀获取现有的key
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}
	s.cc.NewAddress(s.getServices())
	go s.watcher(prefix)
	return s, nil
}

func (s *ServiceDiscovery) Scheme() string {
	panic("implement me")
}

func (s *ServiceDiscovery) ResolveNow(options resolver.ResolveNowOptions) {
	log.Println("ResolveNow")
}

func (s *ServiceDiscovery) Close() {
	log.Println("Close")
	s.cli.Close()
}

// NewServiceDiscovery 新建服务发现
func NewServiceDiscovery(endpoints []string) (resolver.Builder, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &ServiceDiscovery{
		cli: client,
	}, nil
}

// watcher 监听前缀
func (s *ServiceDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
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
}

// SetServiceList 设置新的服务地址
func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serviceList[key] = resolver.Address{Addr: val}
	s.cc.NewAddress(s.getServices())
	log.Println("put key :", key, "val :", val)
}

func (s *ServiceDiscovery) DelServiceList(key string) {
	s.lock.Lock()
	defer s.lock.Lock()
	delete(s.serviceList, key)
	s.cc.NewAddress(s.getServices())
	log.Println("del key:", key)
}

func (s *ServiceDiscovery) getServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, len(s.serviceList))
	for _, address := range s.serviceList {
		addrs = append(addrs, address)
	}
	return addrs
}
