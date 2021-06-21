package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
)

type watcher struct {
	key       string
	ctx       context.Context
	cancel    context.CancelFunc
	watchChan clientv3.WatchChan
	watcher   clientv3.Watcher
	kv        clientv3.KV
	ch        clientv3.WatchChan
}

func newWatcher(ctx context.Context, key string, client *clientv3.Client) *watcher {
	w := &watcher{
		key:     key,
		watcher: clientv3.NewWatcher(client),
	}
	return w
}
