package timer

import "time"

type (
	// Cron is a schedule configured by a cron string.
	cron struct {
		start      time.Time
		sron       string
		executions int32
	}

	// Interval is a schedule configured by a fixed interval.
	interval struct {
		start      time.Time
		interval   int32
		executions int32
	}
)

func (cron) Schedule(initialP progress, task Executer, updateP <-chan progress, stop <-chan int) {

}

func (interval) Schedule(initialP progress, task Executer, updateP <-chan progress, stop <-chan int) {

}
