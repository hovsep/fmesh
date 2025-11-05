package fmesh

import (
	"time"

	"github.com/hovsep/fmesh/cycle"
)

// RuntimeInfo contains information about mesh execution.
type RuntimeInfo struct {
	Cycles    *cycle.Group
	StartedAt time.Time
	StoppedAt time.Time
	Duration  time.Duration
}
