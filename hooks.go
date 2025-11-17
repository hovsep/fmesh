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
// Returns the Hooks registry for method chaining.
func (h *Hooks) BeforeRun(fn func(*FMesh) error) *Hooks {
	h.beforeRun.Add(fn)
	return h
}

// AfterRun registers a hook to be called after the mesh finishes running.
// Returns the Hooks registry for method chaining.
func (h *Hooks) AfterRun(fn func(*FMesh) error) *Hooks {
	h.afterRun.Add(fn)
	return h
}

// CycleBegin registers a hook to be called at the beginning of each cycle.
// Returns the Hooks registry for method chaining.
func (h *Hooks) CycleBegin(fn func(*CycleContext) error) *Hooks {
	h.cycleBegin.Add(fn)
	return h
}

// CycleEnd registers a hook to be called at the end of each cycle.
// Returns the Hooks registry for method chaining.
func (h *Hooks) CycleEnd(fn func(*CycleContext) error) *Hooks {
	h.cycleEnd.Add(fn)
	return h
}
