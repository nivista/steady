package timer

import (
	"time"
)

type schedule func(prog progress, now time.Time) *time.Time
