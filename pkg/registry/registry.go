package registry

import (
	"context"
	"fmt"
	"github.com/crazyfrankie/favorite/internal/config"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.uber.org/zap"
)

type ServiceRegistry struct {
	client     *clientv3.Client
	em         endpoints.Manager
	addr       string
	serviceKey string
	mu         sync.Mutex
	leaseID    clientv3.LeaseID
}

func NewServiceRegistry(cli *clientv3.Client) (*ServiceRegistry, error) {
	addr := "localhost" + config.GetConf().Server.Port
	serviceKey := "service/favorite/" + addr
	em, err := endpoints.NewManager(cli, "service/favorite")
	if err != nil {
		return nil, err
	}

	return &ServiceRegistry{
		client:     cli,
		em:         em,
		addr:       addr,
		serviceKey: serviceKey,
	}, nil
}

func (r *ServiceRegistry) Register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	leaseResp, err := r.client.Grant(ctx, 180)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.leaseID = leaseResp.ID
	r.mu.Unlock()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.em.AddEndpoint(ctx, r.serviceKey,
		endpoints.Endpoint{Addr: r.addr}, clientv3.WithLease(leaseResp.ID)); err != nil {
		return err
	}

	// 开始续约
	go r.KeepAlive()

	return nil
}

func (r *ServiceRegistry) KeepAlive() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		zap.L().Error("KeepAlive failed", zap.Error(err))
	}

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				zap.L().Info("KeepAlive channel closed")
				return
			}
			fmt.Println("Lease renewed")
		case <-ctx.Done():
			return
		}
	}
}

func (r *ServiceRegistry) UnRegister() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := r.em.DeleteEndpoint(ctx, r.serviceKey); err != nil {
		return fmt.Errorf("failed to delete endpoint: %v", err)
	}

	leaseID := r.leaseID

	if _, err := r.client.Revoke(ctx, leaseID); err != nil {
		return fmt.Errorf("failed to revoke lease: %v", err)
	}

	return nil
}
