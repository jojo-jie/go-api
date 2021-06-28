package balance

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

// ServiceRegister 创建租约服务注册
type ServiceRegister struct {
	cli     *clientv3.Client //etcd client
	leaseId clientv3.LeaseID //租约id
	//租约keepalive相应chan
	keepAlive <-chan *clientv3.LeaseKeepAliveResponse
	key       string
	value     string
	ctx       context.Context
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
		ctx:   context.Background(),
	}
	if err := ser.putKeyWithLease(lease); err != nil {
		return nil, err
	}
	return ser, nil
}

// 设置租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	//设置租约时间
	grant, err := s.cli.Grant(s.ctx, lease)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = s.cli.Put(s.ctx, s.key, s.value, clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}
	alive, err := s.cli.KeepAlive(s.ctx, grant.ID)
	if err != nil {
		return err
	}
	s.leaseId = grant.ID
	s.keepAlive = alive
	return nil
}

// ListenLeaseRespChan 监听租约
func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAlive {
		log.Println("续约成功", leaseKeepResp)
	}
	log.Println("关闭续租")
}

// Close 关闭租约
func (s *ServiceRegister) Close() error {
	if _,err:=s.cli.Revoke(s.ctx, s.leaseId);err!=nil {
		return err
	}
	log.Println("撤销租约")
	return s.cli.Close()
}


