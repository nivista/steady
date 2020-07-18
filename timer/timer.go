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
		id        []byte
		progress  progress
		meta      meta
		executer  executer
		scheduler scheduler
	}

	progress struct {
		completed int
		skipped   int
	}

	meta struct {
		creationTime time.Time
	}

	executer interface {
		execute()
		toProto() *common.Task
	}

	scheduler interface {
		schedule(prog progress, now time.Time) (nextFire time.Time, skips int, done bool)
		toProto() *common.Schedule
	}
)

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
func (t *Timer) FromMessageProto(p *messaging.Timer, id []byte) error {
	newT := Timer{id: id}
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
