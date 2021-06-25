package balance

import (
	"github.com/coreos/etcd/clientv3"
	"time"
)

// ServiceRegister 创建租约服务注册
type ServiceRegister struct {
	cli     *clientv3.Client //etcd client
	leaseId clientv3.Lease   //租约id
	//租约keepalive相应chan
	keepAlive <-chan clientv3.LeaseKeepAliveResponse
	key       string
	value     string
}

// NewServiceRegister 新建注册服务
func NewServiceRegister(endpoints []string, serName string, addr string, lease int64) (*ServiceRegister, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	ser := &ServiceRegister{
		cli:   client,
		key:   "/" + schema + "/" + serName + "/" + addr,
		value: addr,
	}

}

func (s *ServiceRegister) put() {

}
