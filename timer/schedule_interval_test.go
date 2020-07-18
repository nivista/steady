package timer

import (
	"testing"
	"time"
)

func TestIntervalSchedule(t *testing.T) {
	type testCase struct {
		interval         interval
		progress         progress
		time             time.Time
		expectedNextFire time.Time
		expectedSkips    int
		expectedDone     bool
	}

	testCases := []testCase{
		{
			interval: interval{
				start:      time.Unix(0, 0),
				interval:   10,
				executions: 4,
			},
			progress:         progress{skipped: 1},
			time:             time.Unix(0, 0),
			expectedNextFire: time.Unix(10, 0),
			expectedSkips:    0,
			expectedDone:     false,
		},
		{
			interval: interval{
				start:      time.Unix(0, 0),
				interval:   10,
				executions: 4,
			},
			progress: progress{
				skipped:   1,
				completed: 4,
			},
			time:         time.Unix(0, 0),
			expectedDone: true,
		},
		{
			interval: interval{
				start:      time.Date(2012, time.April, 0, 0, 0, 0, 132, time.UTC),
				interval:   100,
				executions: 4,
			},
			progress: progress{
				skipped:   2,
				completed: 3,
			},
			time:             time.Date(2012, time.April, 0, 0, 0, 0, 131, time.UTC).Add(600e9),
			expectedNextFire: time.Date(2012, time.April, 0, 0, 0, 0, 131, time.UTC).Add(600e9),
		},
		{
			interval: interval{
				start:      time.Date(2012, time.March, 20, 0, 0, 0, 0, time.UTC),
				interval:   60,
				executions: -1,
			},
			progress: progress{
				skipped:   2,
				completed: 30,
			},
			time:             time.Date(2012, time.March, 20, 0, 35, 0, 0, time.UTC),
			expectedNextFire: time.Date(2012, time.March, 20, 0, 35, 0, 0, time.UTC),
			expectedSkips:    3,
		},
		{
			interval: interval{
				start:      time.Date(2012, time.March, 20, 0, 0, 0, 0, time.UTC),
				interval:   60,
				executions: -1,
			},
			progress: progress{
				skipped:   2,
				completed: 30,
			},
			time:             time.Date(2012, time.March, 20, 0, 30, 0, 0, time.UTC),
			expectedNextFire: time.Date(2012, time.March, 20, 0, 32, 0, 0, time.UTC),
		},
	}

	for idx, test := range testCases {
		nextFire, skips, done := test.interval.schedule(test.progress, test.time)
		if done != test.expectedDone {
			t.Errorf("test at idx %v: expected '%v' for done, got '%v'", idx, test.expectedDone, done)
		}

		if skips != test.expectedSkips {
			t.Errorf("test at idx %v: expected '%v' for skips, got '%v'", idx, test.expectedSkips, skips)
		}

		if nextFire != test.expectedNextFire {
			t.Errorf("test at idx %v: expected '%v' for nextFire, got '%v'", idx, test.expectedNextFire, nextFire)
		}
	}
}
