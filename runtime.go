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
}

func NewRuntimeInfo() *RuntimeInfo {
	return &RuntimeInfo{
		Cycles: cycle.NewGroup(),
	}
}

func (r *RuntimeInfo) MarkStarted() {
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now()
	}
}

func (r *RuntimeInfo) MarkStopped() {
	if r.StoppedAt.IsZero() {
		r.StoppedAt = time.Now()
	}
}

func (r *RuntimeInfo) Duration() time.Duration {
	if r.StartedAt.IsZero() {
		return 0
	}
	if r.StoppedAt.IsZero() {
		return time.Since(r.StartedAt)
	}

	return r.StoppedAt.Sub(r.StartedAt)
}
