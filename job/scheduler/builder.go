package scheduler

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/crazyfrankie/favorite/job"
)

type CronJobBuilder struct {
	metric *prometheus.SummaryVec
	log    *zap.Logger
}

func NewCronJobBuilder(l *zap.Logger) *CronJobBuilder {
	metric := prometheus.NewSummaryVec(prometheus.SummaryOpts{}, []string{"name", "success"})
	prometheus.MustRegister(metric)

	return &CronJobBuilder{
		metric: metric,
		log:    l,
	}
}

func (c *CronJobBuilder) Builder(j job.Job) cron.Job {
	name := j.Name()

	return cronJob(func() error {
		start := time.Now()
		c.log.Debug("任务开始",
			zap.String("name", name),
			zap.String("time", start.String()))

		var success bool
		defer func() {
			duration := time.Since(start).Milliseconds()
			c.log.Debug("任务结束",
				zap.String("name", name))
			c.metric.WithLabelValues(name,
				strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := j.Run()
		success = err == nil
		if err != nil {
			c.log.Error("任务执行失败",
				zap.String("name", name),
				zap.Error(err))
		}

		return nil
	})
}

type cronJob func() error

func (c cronJob) Run() {
	_ = c()
}
