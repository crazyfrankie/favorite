package scheduler

import (
	"context"

	"github.com/crazyfrankie/favorite/internal/biz/service"
)

type DataScheduler struct {
	opt *option
	svc *service.FavoriteServer
}

func NewScheduler(svc *service.FavoriteServer, opts ...Option) *DataScheduler {
	opt := &option{
		timeout: 60,
	}
	for _, o := range opts {
		o(opt)
	}

	return &DataScheduler{
		opt: opt,
		svc: svc,
	}
}

func (s *DataScheduler) Name() string {
	return "data_sync"
}

func (s *DataScheduler) Run() error {
	_, cancel := context.WithTimeout(context.Background(), s.opt.timeout)
	defer cancel()
	if s.opt.threshold > 0 {
		// 默认不采用阈值方式, 先校验, 如果设置了就走这条路
		// 根据阈值判断是否持久化
		// 先获取当前点赞数
	}

	return nil
}
