package fmesh

import (
	"time"

	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/hook"
)

// RuntimeInfo contains information about mesh execution.
type RuntimeInfo struct {
	Cycles    *cycle.Group
	StartedAt time.Time
	StoppedAt time.Time
	Duration  time.Duration
}

// CycleContext provides context for cycle-level hooks.
type CycleContext struct {
	FMesh *FMesh
	Cycle *cycle.Cycle
}

// Hooks is a registry of all hook types for FMesh.
// All hooks are stored in typed groups and executed in insertion order.
type Hooks struct {
	beforeRun  *hook.Group[*FMesh]
	afterRun   *hook.Group[*FMesh]
	cycleBegin *hook.Group[*CycleContext]
	cycleEnd   *hook.Group[*CycleContext]
}
