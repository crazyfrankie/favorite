package scheduler

type MonitorScheduler struct {
	// 监控间隔
	interval int
}

func NewMonitorScheduler(interval int) *MonitorScheduler {
	return &MonitorScheduler{interval: interval}
}

func (m MonitorScheduler) Name() string {
	return "system_monitor"
}

// Run 监控业务系统的健康程度, 灵活切换数据同步方式
func (m MonitorScheduler) Run() error {
	//TODO implement me
	panic("implement me")
}
