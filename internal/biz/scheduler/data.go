package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/crazyfrankie/favorite/internal/biz/repository"
)

type DataScheduler struct {
	opt  *option
	cron *cron.Cron
	repo *repository.FavoriteRepo
}

func NewScheduler(repo *repository.FavoriteRepo, opts ...Option) *DataScheduler {
	opt := &option{
		interval: 120, // 默认 120 分钟同步
	}
	for _, o := range opts {
		o(opt)
	}

	cr := cron.New(cron.WithSeconds())

	return &DataScheduler{
		opt:  opt,
		cron: cr,
		repo: repo,
	}
}

// TODO

func (s *DataScheduler) SyncToDB(ctx context.Context) error {
	if s.opt.threshold > 0 {
		// 默认不采用阈值方式, 先校验, 如果设置了就走这条路
		// 根据阈值判断是否持久化
		// 先获取当前点赞数
	}

	if s.opt.interval > 0 {
		// 根据设置的定时任务进行持久化
		crStr := fmt.Sprintf("@every %dm", s.opt.interval)
		_, err := s.cron.AddJob(crStr, &IntervalJob{interval: s.opt.interval})
		if err != nil {
			return err
		}
		s.cron.Start()
	}

	return nil
}

type ThresholdJob struct {
	threshold int64
}

func (t *ThresholdJob) Run() {

}

type IntervalJob struct {
	interval time.Duration
}

func (s *IntervalJob) Run() {
	//TODO implement me
	panic("implement me")
}
