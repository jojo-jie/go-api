package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"tag-service/pkg/registry"
)

var _ registry.Watcher = &watcher{}

type watcher struct {
	key       string
	ctx       context.Context
	cancel    context.CancelFunc
	watchChan clientv3.WatchChan
	watcher   clientv3.Watcher
	kv        clientv3.KV
	ch        clientv3.WatchChan
}

func newWatcher(ctx context.Context, key string, client *clientv3.Client) (*watcher, error) {
	w := &watcher{
		key:     key,
		watcher: clientv3.NewWatcher(client),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.watchChan = w.watcher.Watch(w.ctx, key, clientv3.WithPrefix(), clientv3.WithRev(0))
	err := w.watcher.RequestProgress(context.Background())
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	for {
		select {
		case <-w.ctx.Done():
			return nil, w.ctx.Err()
		case <-w.watchChan:
		}
		resp, err := w.kv.Get(w.ctx, w.key, clientv3.WithPrefix())
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
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.watcher.Close()
}
