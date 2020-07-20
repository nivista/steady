package timer

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/protobuf/proto"
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
		completedExecutions int
		lastExecution       int
	}

	meta struct {
		creationTime time.Time
	}

	executer interface {
		execute()
		toProto() *common.Task
	}

	scheduler interface {
		schedule(prog progress, now time.Time) (nextFire time.Time, executionNumber int, done bool)
		toProto() *common.Schedule
	}
)

// Context manages the goroutines created by work from a group of timers.
type (
	Context interface {
		Cancel()
		done() <-chan struct{}
	}

	context struct {
		ch  chan struct{}
		mux sync.Mutex
	}
)

// NewContext returns an initialized timerContext.
func NewContext() Context {
	return &context{ch: make(chan struct{})}
}

func (w *context) Cancel() {
	w.mux.Lock()
	defer w.mux.Unlock()

	select {
	case <-w.ch:
	default:
		close(w.ch)
	}
}

func (w *context) done() <-chan struct{} {
	return w.ch
}

// Run starts the timers execution
func (t *Timer) Run(ctx Context, consumer chan<- *sarama.ProducerMessage, clock clockwork.Clock) {

	pb := t.ToMessageProto()

	for {
		nextFire, executionNumber, done := t.scheduler.schedule(t.progress, clock.Now())

		if done {
			return
		}

		select {
		case <-clock.After(nextFire.Sub(clock.Now())):
			t.executer.execute()
			pb.Progress.LastExecution = int32(executionNumber)
			pb.Progress.CompletedExecutions++

			val, err := proto.Marshal(pb)
			if err != nil {
				panic(fmt.Sprint("timer Run :", err))
			}

			consumer <- &sarama.ProducerMessage{
				Topic: "timer",
				Key:   sarama.ByteEncoder(t.id),
				Value: sarama.ByteEncoder(val),
			}

		case <-ctx.done():
			return
		}
	}
}

// ToMessageProto creates a messaging.Timer from this Timer.
func (t *Timer) ToMessageProto() *messaging.Timer {
	p := messaging.Timer{}
	p.Meta = &common.Meta{CreateTime: timestamppb.New(t.meta.creationTime)}

	p.Progress = &common.Progress{
		CompletedExecutions: int32(t.progress.completedExecutions),
		LastExecution:       int32(t.progress.lastExecution),
	}

	p.Task = t.executer.toProto()
	p.Schedule = t.scheduler.toProto()
	return &p
}

// FromMessageProto gets a Timer from a messaging.Timer.
func (t *Timer) FromMessageProto(p *messaging.Timer, id []byte) error {
	newT := Timer{id: id}
	newT.meta = meta{creationTime: p.Meta.CreateTime.AsTime()}

	newT.progress = progress{
		completedExecutions: int(p.Progress.CompletedExecutions),
		lastExecution:       int(p.Progress.LastExecution),
	}

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
