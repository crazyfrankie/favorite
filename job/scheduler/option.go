package scheduler

import "time"

type option struct {
	// 阈值
	threshold int64
	// 调度超时时间
	timeout time.Duration
}

type Option func(*option)

func WithThreshold(threshold int64) Option {
	return func(o *option) {
		o.threshold = threshold
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *option) {
		o.timeout = timeout
	}
}
