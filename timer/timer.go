package timer

import (
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	// Timer is a recurring task that will invoke handlers when recurring tasks execute or are finished.
	Timer interface {
		// WithProgress should return a new timer with the given Progress.
		WithProgress(pb *messaging.Progress) Timer

		// Start starts the timers execution. It should be a no-op if the timer is active.
		Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), clock clockwork.Clock)

		// Stop should stop timer execution synchronously.
		Stop()
	}

	// defining type here rather than using only protobuf Progress type, this can be safely copied
	progress struct {
		completedExecutions int32
		lastExecution       time.Time
	}
)
