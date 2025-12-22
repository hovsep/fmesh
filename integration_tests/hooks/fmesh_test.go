package hooks

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooks_AllTypes(t *testing.T) {
	// Track hook execution
	var executionLog []string

	// Create a simple component
	c := component.New("processor").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	// Create a mesh with all hook types
	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				executionLog = append(executionLog, "beforeRun")
				return nil
			})

			h.AfterRun(func(fm *fmesh.FMesh) error {
				executionLog = append(executionLog, "afterRun")
				return nil
			})

			h.CycleBegin(func(ctx *fmesh.CycleContext) error {
				executionLog = append(executionLog, "cycleBegin")
				return nil
			})

			h.CycleEnd(func(ctx *fmesh.CycleContext) error {
				executionLog = append(executionLog, "cycleEnd")
				return nil
			})
		})

	// Add initial input
	fm.ComponentByName("processor").InputByName("in").PutSignals(signal.New(1))

	// Run mesh
	_, err := fm.Run()
	require.NoError(t, err)

	// Verify the exact execution order: beforeRun -> cycles -> afterRun
	// Cycle hooks fire twice: once for processing, once for completion
	assert.Equal(t, []string{
		"beforeRun",
		"cycleBegin",
		"cycleEnd",
		"cycleBegin",
		"cycleEnd",
		"afterRun",
	}, executionLog)
}

func TestHooks_CycleContext(t *testing.T) {
	cycleNumbers := []int{}

	// Create a simple component
	c := component.New("processor").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.CycleBegin(func(ctx *fmesh.CycleContext) error {
				cycleNumbers = append(cycleNumbers, ctx.Cycle.Number())
				return nil
			})
		})

	// Add input for multiple cycles
	fm.ComponentByName("processor").InputByName("in").
		PutSignals(signal.New(1), signal.New(2), signal.New(3))

	// Run mesh
	_, err := fm.Run()
	require.NoError(t, err)

	// Verify cycle numbers are sequential
	// Cycle 1: process all signals, Cycle 2: mesh completes
	assert.Equal(t, []int{1, 2}, cycleNumbers)
}

func TestHooks_MultipleHooksPerType(t *testing.T) {
	// Test that multiple hooks of the same type execute in order
	var log []string

	c := component.New("test").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "first")
				return nil
			})

			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "second")
				return nil
			})

			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "third")
				return nil
			})
		})

	fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1))

	_, err := fm.Run()
	require.NoError(t, err)

	// Verify all hooks executed in order
	assert.Equal(t, []string{"first", "second", "third"}, log)
}

func TestHooks_ContextAccess(t *testing.T) {
	var meshName string
	var cycleNumber int
	var activationCount int

	c := component.New("test").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	fm := fmesh.New("my-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.CycleEnd(func(ctx *fmesh.CycleContext) error {
				// Access both FMesh and Cycle through context
				meshName = ctx.FMesh.Name()
				cycleNumber = ctx.Cycle.Number()
				activationCount = ctx.Cycle.ActivationResults().Len()
				return nil
			})
		})

	fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1))

	_, err := fm.Run()
	require.NoError(t, err)

	// Verify CycleContext provides access to both FMesh and Cycle
	// Values are from the last cycle (completion cycle)
	assert.Equal(t, "my-mesh", meshName)
	assert.Equal(t, 2, cycleNumber)
	assert.Equal(t, 1, activationCount)
}

func TestHooks_FireOncePerCycle(t *testing.T) {
	var beginCount int
	var endCount int

	c := component.New("processor").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.CycleBegin(func(ctx *fmesh.CycleContext) error {
				beginCount++
				return nil
			})
			h.CycleEnd(func(ctx *fmesh.CycleContext) error {
				endCount++
				return nil
			})
		})

	// Add multiple signals to trigger multiple cycles
	fm.ComponentByName("processor").InputByName("in").
		PutSignals(signal.New(1), signal.New(2), signal.New(3))

	runtimeInfo, err := fm.Run()
	require.NoError(t, err)

	// Verify hooks fire exactly once per cycle
	actualCycles := runtimeInfo.Cycles.Len()
	assert.Equal(t, actualCycles, beginCount, "CycleBegin should fire once per cycle")
	assert.Equal(t, actualCycles, endCount, "CycleEnd should fire once per cycle")
	assert.Equal(t, beginCount, endCount, "Begin and End should fire same number of times")
}

func TestHooks_RunWithError(t *testing.T) {
	var beforeRunFired bool
	var afterRunFired bool

	// Create a mesh with chainable error (simulating previous failed run)
	fm := fmesh.New("test-mesh").
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				beforeRunFired = true
				return nil
			})
			h.AfterRun(func(fm *fmesh.FMesh) error {
				afterRunFired = true
				return nil
			})
		}).
		WithChainableErr(assert.AnError) // Simulate previous error

	_, err := fm.Run()
	require.Error(t, err)

	// Run() returns immediately when mesh has chainable error - hooks don't fire
	assert.False(t, beforeRunFired, "BeforeRun should not fire when mesh has chainable error")
	assert.False(t, afterRunFired, "AfterRun should not fire when mesh has chainable error")
}

func TestHooks_EmptyMesh(t *testing.T) {
	var beforeRunFired bool
	var afterRunFired bool

	// Create mesh with no components - this will error
	fm := fmesh.New("empty-mesh").
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				beforeRunFired = true
				return nil
			})
			h.AfterRun(func(fm *fmesh.FMesh) error {
				afterRunFired = true
				return nil
			})
		})

	_, err := fm.Run()
	require.Error(t, err) // Empty mesh returns error

	// BeforeRun and AfterRun should still fire (defer behavior)
	assert.True(t, beforeRunFired)
	assert.True(t, afterRunFired)
}

func TestHooks_MultipleSetupCalls(t *testing.T) {
	var log []string

	c := component.New("test").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	// Multiple SetupHooks calls should accumulate hooks
	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "first-setup")
				return nil
			})
		}).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "second-setup")
				return nil
			})
		}).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) error {
				log = append(log, "third-setup")
				return nil
			})
		})

	fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1))

	_, err := fm.Run()
	require.NoError(t, err)

	// All hooks from all SetupHooks calls should execute in order
	assert.Equal(t, []string{"first-setup", "second-setup", "third-setup"}, log)
}
