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

// newRuntimeInfo constructor. historyLimit caps how many past cycles are
// retained in Cycles (0 means unlimited); the group itself enforces the cap.
func newRuntimeInfo(historyLimit int) *RuntimeInfo {
	return &RuntimeInfo{
		Cycles: cycle.NewGroup().SetLenLimit(historyLimit),
	}
}

// MarkStarted sets when fmesh is started running.
func (r *RuntimeInfo) MarkStarted() {
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now()
	}
}

// MarkStopped sets when the fmesh is stopped running.
func (r *RuntimeInfo) MarkStopped() {
	if r.StoppedAt.IsZero() {
		r.StoppedAt = time.Now()
	}
}

// Duration returns the duration of the fmesh execution.
func (r *RuntimeInfo) Duration() time.Duration {
	if r.StartedAt.IsZero() {
		return 0
	}
	if r.StoppedAt.IsZero() {
		return time.Since(r.StartedAt)
	}

	return r.StoppedAt.Sub(r.StartedAt)
}
