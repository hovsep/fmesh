package fmesh

import (
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/hook"
)

// CycleContext provides context for cycle-level hooks.
type CycleContext struct {
	FMesh *FMesh
	Cycle *cycle.Cycle
}

// Hooks is a registry of all hook types for FMesh.
type Hooks struct {
	beforeRun  *hook.Group[*FMesh]
	afterRun   *hook.Group[*FMesh]
	cycleBegin *hook.Group[*CycleContext]
	cycleEnd   *hook.Group[*CycleContext]
}

// NewHooks creates a new hooks registry with empty hook groups.
func NewHooks() *Hooks {
	return &Hooks{
		beforeRun:  hook.NewGroup[*FMesh](),
		afterRun:   hook.NewGroup[*FMesh](),
		cycleBegin: hook.NewGroup[*CycleContext](),
		cycleEnd:   hook.NewGroup[*CycleContext](),
	}
}

// BeforeRun registers a hook to be called before the mesh starts running.
func (h *Hooks) BeforeRun(fn func(*FMesh)) {
	h.beforeRun.Add(fn)
}

// AfterRun registers a hook to be called after the mesh finishes running.
func (h *Hooks) AfterRun(fn func(*FMesh)) {
	h.afterRun.Add(fn)
}

// CycleBegin registers a hook to be called at the beginning of each cycle.
func (h *Hooks) CycleBegin(fn func(*CycleContext)) {
	h.cycleBegin.Add(fn)
}

// CycleEnd registers a hook to be called at the end of each cycle.
func (h *Hooks) CycleEnd(fn func(*CycleContext)) {
	h.cycleEnd.Add(fn)
}
