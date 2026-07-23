package fmesh

import (
	"fmt"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/internal/hook"
	"github.com/hovsep/fmesh/port"
)

// CycleContext provides context for cycle-level hooks.
type CycleContext struct {
	FMesh *FMesh
	Cycle *cycle.Cycle
}

// ComponentAddedContext provides context when a component is added to the mesh.
type ComponentAddedContext struct {
	FMesh     *FMesh
	Component *component.Component
}

// Hooks is a registry of all hook types for FMesh.
type Hooks struct {
	onComponentAdded *hook.Group[*ComponentAddedContext]
	beforeRun        *hook.Group[*FMesh]
	afterRun         *hook.Group[*FMesh]
	beforeCycle      *hook.Group[*CycleContext]
	afterCycle       *hook.Group[*CycleContext]
}

// newHooks creates a new hooks registry with default hooks.
func newHooks() *Hooks {
	return &Hooks{
		onComponentAdded: hook.NewGroup[*ComponentAddedContext](),
		beforeRun:        hook.NewGroup[*FMesh]().Add(getDefaultBeforeRunHook()),
		afterRun:         hook.NewGroup[*FMesh](),
		beforeCycle:      hook.NewGroup[*CycleContext](),
		afterCycle:       hook.NewGroup[*CycleContext](),
	}
}

// getDefaultBeforeRunHook validates the mesh structure on every run,
// so components added between runs are validated too.
func getDefaultBeforeRunHook() func(*FMesh) error {
	return validateMeshStructure
}

// validateMeshStructure validates components in name order so validation
// errors are deterministic.
func validateMeshStructure(fm *FMesh) error {
	for _, c := range fm.Components().AllOrdered() {
		if err := validateComponentStructure(fm, c); err != nil {
			return err
		}
	}
	return nil
}

func validateComponentStructure(fm *FMesh, c *component.Component) error {
	if c.ParentMesh() != fm {
		return fmt.Errorf("component %q has wrong parent mesh", c.Name())
	}
	return c.Outputs().ForEach(func(p *port.Port) error {
		if p.ParentComponent() != c {
			return fmt.Errorf("output port %q has wrong parent component in component %q", p.Name(), c.Name())
		}
		return p.Pipes().ForEach(func(dest *port.Port) error {
			parent := dest.ParentComponent()
			destComponent, ok := parent.(*component.Component)
			if !ok || destComponent == nil {
				return fmt.Errorf("destination port %q has invalid parent component", dest.Name())
			}
			if destComponent.ParentMesh() != fm {
				return fmt.Errorf("destination component %q belongs to a different mesh", destComponent.Name())
			}
			return nil
		})
	})
}

// OnComponentAdded registers a hook called after each component is successfully added to the mesh.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnComponentAdded(fn func(*ComponentAddedContext) error) *Hooks {
	h.onComponentAdded.Add(fn)
	return h
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

// BeforeCycle registers a hook to be called at the beginning of each cycle.
// Returns the Hooks registry for method chaining.
func (h *Hooks) BeforeCycle(fn func(*CycleContext) error) *Hooks {
	h.beforeCycle.Add(fn)
	return h
}

// AfterCycle registers a hook to be called at the end of each cycle.
// Returns the Hooks registry for method chaining.
func (h *Hooks) AfterCycle(fn func(*CycleContext) error) *Hooks {
	h.afterCycle.Add(fn)
	return h
}
