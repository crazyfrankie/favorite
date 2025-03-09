package scheduler

import "time"

type option struct {
	threshold int64
	// 默认以分钟为单位
	interval time.Duration
}

type Option func(*option)

func WithThreshold(threshold int64) Option {
	return func(o *option) {
		o.threshold = threshold
	}
}

func WithTickTime(interval time.Duration) Option {
	return func(o *option) {
		o.interval = interval
	}
}
