package scheduler

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type MonitorScheduler struct {
	cron *cron.Cron
}

func NewMonitorScheduler() *MonitorScheduler {
	cr := cron.New(cron.WithSeconds())

	return &MonitorScheduler{cron: cr}
}

// TODO

func (m *MonitorScheduler) MonitorBusinessVolume(ctx context.Context) {
	_, err := m.cron.AddFunc("@every 20s", Job)
	if err != nil {
		zap.L().Warn("failed to start cron job", zap.Error(err))
		for i := 0; i < 5; i++ {
			if _, err := m.cron.AddFunc("@every 20s", Job); err == nil {
				break
			}
		}
	}
}

// Job 监控业务量, 灵活切换点赞同步方式
func Job() {

}
