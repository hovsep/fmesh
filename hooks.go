package fmesh

import "github.com/hovsep/fmesh/cycle"

// CycleContext provides context for cycle-level hooks.
type CycleContext struct {
	FMesh *FMesh
	Cycle *cycle.Cycle
}

// Hooks is a registry of all hook types for FMesh.
// All hooks are stored in typed groups and executed in insertion order.
type Hooks struct {
	beforeRun  *HookGroup[*FMesh]
	afterRun   *HookGroup[*FMesh]
	cycleBegin *HookGroup[*CycleContext]
	cycleEnd   *HookGroup[*CycleContext]
}

// NewHooks creates a new hooks registry with empty hook groups.
func NewHooks() *Hooks {
	return &Hooks{
		beforeRun:  NewHookGroup[*FMesh](),
		afterRun:   NewHookGroup[*FMesh](),
		cycleBegin: NewHookGroup[*CycleContext](),
		cycleEnd:   NewHookGroup[*CycleContext](),
	}
}

// BeforeRun registers a hook to be called before the mesh starts running.
func (h *Hooks) BeforeRun(hook func(*FMesh)) {
	h.beforeRun.Add(hook)
}

// AfterRun registers a hook to be called after the mesh finishes running.
func (h *Hooks) AfterRun(hook func(*FMesh)) {
	h.afterRun.Add(hook)
}

// CycleBegin registers a hook to be called at the beginning of each cycle.
func (h *Hooks) CycleBegin(hook func(*CycleContext)) {
	h.cycleBegin.Add(hook)
}

// CycleEnd registers a hook to be called at the end of each cycle.
func (h *Hooks) CycleEnd(hook func(*CycleContext)) {
	h.cycleEnd.Add(hook)
}
