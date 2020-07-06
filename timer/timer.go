package timer

import (
	"errors"
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Timer definition
type (
	Timer struct {
		progress progress
		meta     meta
		executer
		scheduler
	}

	meta struct {
		creationTime time.Time
	}

	progress struct {
		completed int
		skipped   int
	}

	executer interface {
		execute()
		toProto() *common.Task
	}

	scheduler interface {
		schedule(progress *progress, task executer, updateProgress chan<- int, stop <-chan int)
		toProto() *common.Schedule
	}
)

// Work returns a channel that counts the number of times progress was updated, as well as a stop channel.
func (t *Timer) Work() (updateProgress chan<- int, stop <-chan int) {
	updateProgress = make(chan int)
	stop = make(chan int)
	go t.schedule(&t.progress, t.executer, updateProgress, stop)
	return
}

// ToMessageProto creates a messaging.Timer from this Timer.
func (t *Timer) ToMessageProto() *messaging.Timer {
	p := messaging.Timer{}
	p.Meta = &common.Meta{CreateTime: timestamppb.New(t.meta.creationTime)}
	p.Progress = &common.Progress{Completed: int32(t.progress.completed), Skipped: int32(t.progress.skipped)}
	p.Task = t.executer.toProto()
	p.Schedule = t.scheduler.toProto()
	return &p
}

// FromMessageProto gets a Timer from a messaging.Timer.
func (t *Timer) FromMessageProto(p *messaging.Timer) error {
	newT := Timer{}
	newT.meta = meta{creationTime: p.Meta.CreateTime.AsTime()}
	newT.progress = progress{completed: int(p.Progress.Completed), skipped: int(p.Progress.Skipped)}

	switch task := p.Task.Task.(type) {
	case *common.Task_HttpConfig:
		var h http
		if err := h.fromProto(task); err != nil {
			return err
		}
		newT.executer = h
	default:
		return errors.New("Unknown Task")
	}

	switch sched := p.Schedule.Schedule.(type) {
	case *common.Schedule_CronConfig:
		var c cron
		if err := c.fromProto(sched); err != nil {
			return err
		}
		newT.scheduler = c
	case *common.Schedule_IntervalConfig:
		var i interval
		if err := i.fromProto(sched); err != nil {
			return err
		}
		newT.scheduler = i
	default:
		return errors.New("Unknown Schedule")
	}

	*t = newT
	return nil
}
