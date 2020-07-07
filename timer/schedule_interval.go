package timer

import (
	"fmt"
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// interval is a scheduler configured by a fixed interval.
type interval struct {
	start      time.Time
	interval   int
	executions int
}

func (i interval) schedule(progress *progress, task executer, updateProgress chan<- int, stop <-chan int) {
	defer close(updateProgress)

	intervalNano := i.interval * 1e9

	if progress.completed >= i.executions {
		<-stop
		return
	}
	fmt.Println(time.Now())
Catchup:

	intervalsAhead := progress.completed + progress.skipped
	nextFire := i.start.Add(time.Duration(intervalNano * intervalsAhead))
	diff := time.Now().Sub(nextFire)

	// next fire is in the past
	if diff > 0 {
		// fire once now
		task.execute()

		progress.completed++

		if progress.completed >= i.executions {
			fmt.Println("here")
			return
		}

		select {
		case <-stop:
			fmt.Println("stopped while executing")
			return
		default:
			updateProgress <- 0
		}

		intervalsAhead = progress.completed + progress.skipped
		nextFire = i.start.Add(time.Duration(intervalNano * intervalsAhead))
		diff = time.Now().Sub(nextFire)

		// for every other missed fire just skip
		for diff > 0 {
			fmt.Println("skipping")
			progress.skipped++

			select {
			case <-stop:
				fmt.Println("stopped while skipping")
				return
			default:
				updateProgress <- 0
				fmt.Println("skipping2")

			}

			intervalsAhead = progress.completed + progress.skipped
			nextFire = i.start.Add(time.Duration(intervalNano * intervalsAhead))
			diff = time.Now().Sub(nextFire)
		}
	}
	fmt.Println(nextFire)

	t := <-time.NewTimer(time.Until(nextFire)).C
	fmt.Println(t)
	goto Catchup

}

func (i interval) toProto() *common.Schedule {
	return &common.Schedule{
		Schedule: &common.Schedule_IntervalConfig{
			IntervalConfig: &common.IntervalConfig{
				StartTime:  timestamppb.New(i.start),
				Interval:   int32(i.interval),
				Executions: int32(i.executions),
			},
		},
	}
}

func (i *interval) fromProto(p *common.Schedule_IntervalConfig) error {
	intConfig := p.IntervalConfig

	*i = interval{
		start:      intConfig.StartTime.AsTime(),
		interval:   int(intConfig.Interval),
		executions: int(intConfig.Executions),
	}

	return nil
}
