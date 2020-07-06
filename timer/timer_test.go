package timer

import (
	"reflect"
	"testing"
	"time"
)

func getHttpCron() *Timer {
	return &Timer{
		meta: meta{
			creationTime: time.Unix(0, 0).UTC(),
		},
		progress: progress{
			completed: 5,
			skipped:   1,
		},
		executer: http{
			method:  GET,
			url:     "http://example.com",
			body:    "",
			headers: map[string]string{"set-cookie": "hello:cookie"},
		},
		scheduler: cron{
			start:      time.Unix(0, 0).UTC(),
			executions: 10,
			min:        -1,
			hour:       2,
			dayOfMonth: 3,
			month:      3,
			dayOfWeek:  -1,
		},
	}
}

func getHttpInterval() *Timer {
	return &Timer{
		meta: meta{
			creationTime: time.Unix(0, 0).UTC(),
		},
		progress: progress{
			completed: 5,
			skipped:   1,
		},
		executer: http{
			method:  GET,
			url:     "http://example.com",
			body:    "",
			headers: map[string]string{"set-cookie": "hello:cookie"},
		},
		scheduler: interval{
			start:      time.Unix(0, 0).UTC(),
			interval:   100,
			executions: 10,
		},
	}
}

func TestProto(t *testing.T) {
	timers := []*Timer{getHttpCron(), getHttpInterval()}

	for idx, timer := range timers {
		proto := timer.ToMessageProto()

		var timerTwo Timer
		if err := timerTwo.FromMessageProto(proto); err != nil {
			t.Errorf("FromMessageProto at idx %v failed w/ error: %v\n", idx, err)
		}

		if !reflect.DeepEqual(*timer, timerTwo) {
			t.Errorf("Going to and from proto yielded a different timer at idx %v\n", idx)
		}
	}
}

type MockExecutor struct {
	count int
}

func (m *MockExecutor) execute() {
	m.count++
}
