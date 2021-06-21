package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"tag-service/pkg/registry"
	"time"
)

var (
	_ registry.Registrar = &Registry{}
	_ registry.Discovery = &Registry{}
)

// Option is etcd registry option

type Option func(o *options)

type options struct {
	ctx       context.Context
	namespace string
	ttl       time.Duration
}

func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func Namespace(ns string) Option {
	return func(o *options) {
		o.namespace = ns
	}
}

func RegisterTTl(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

type Registry struct {
	opts   *options
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func New(client *clientv3.Client, opts ...Option) (r *Registry) {
	options := &options{
		ctx:       context.Background(),
		namespace: "/microservices",
		ttl:       time.Second * 15,
	}

	for _, o := range opts {
		o(options)
	}

	return &Registry{
		opts:   options,
		client: client,
		kv:     clientv3.NewKV(client),
	}
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.ID)
	value, err := marshal(service)
	if err != nil {
		return err
	}
	if r.lease != nil {
		r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	grant, err := r.lease.Grant(ctx, int64(r.opts.ttl.Seconds()))
	if err != nil {
		return err
	}
	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}

	// 定期续租
	hb, err := r.client.KeepAlive(ctx, grant.ID)
	if err != nil {
		return err
	}
	log.Printf("registry %d\n", grant.ID)
	go func() {
		for {
			select {
			case _, ok := <-hb:
				if !ok {
					log.Printf("续约失败 %d\n", grant.ID)
					return
				}
			case <-r.opts.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	defer func() {
		if r.lease != nil {
			r.lease.Close()
		}
	}()
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

//GetService return the service instances in memory according to the service name.
func (r *Registry) GetService(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", r.opts.namespace, name)
	resp, err := r.kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var items []*registry.ServiceInstance
	for _, kv := range resp.Kvs {
		si, err := unmarshal(kv.Value)
		if err != nil {
			return nil, err
		}
		items = append(items, si)
	}
	return items, nil
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	panic("implement me")
}
