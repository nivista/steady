package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
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
