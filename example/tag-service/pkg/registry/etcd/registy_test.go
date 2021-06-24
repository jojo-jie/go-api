package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"tag-service/pkg/registry"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	conf := clientv3.Config{}
	conf.Endpoints = []string{"127.0.0.1:2379"}
	client, err := clientv3.New(conf)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	s := &registry.ServiceInstance{
		ID:   "0",
		Name: "helloworld",
	}

	r := New(client)
	w, err := r.Watch(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	go func() {
		for {
			res, err := w.Next()
			if err != nil {
				return
			}
			t.Logf("watch: %d", len(res))
			for _, r := range res {
				t.Logf("next: %+v", r)
			}
		}
	}()
	time.Sleep(time.Second)

	if err := r.Register(ctx, s); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	res, err := r.GetService(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 && res[0].Name != s.Name {
		t.Errorf("not expected: %+v", res)
	}

	if err := r.Deregister(ctx, s); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	res, err = r.GetService(ctx, s.Name)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Errorf("not expected empty")
	}
}
