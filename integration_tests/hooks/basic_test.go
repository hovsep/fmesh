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

	// Create simple component
	c := component.New("processor").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	// Create mesh with all hook types
	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) {
				executionLog = append(executionLog, "beforeRun")
			})

			h.AfterRun(func(fm *fmesh.FMesh) {
				executionLog = append(executionLog, "afterRun")
			})

			h.CycleBegin(func(ctx *fmesh.CycleContext) {
				executionLog = append(executionLog, "cycleBegin")
			})

			h.CycleEnd(func(ctx *fmesh.CycleContext) {
				executionLog = append(executionLog, "cycleEnd")
			})
		})

	// Add initial input
	fm.ComponentByName("processor").InputByName("in").PutSignals(signal.New(1))

	// Run mesh
	_, err := fm.Run()
	require.NoError(t, err)

	// Verify exact execution order: beforeRun -> cycles -> afterRun
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

	// Create simple component
	c := component.New("processor").
		AddInputs("in").
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	fm := fmesh.New("test-mesh").
		AddComponents(c).
		SetupHooks(func(h *fmesh.Hooks) {
			h.CycleBegin(func(ctx *fmesh.CycleContext) {
				cycleNumbers = append(cycleNumbers, ctx.Cycle.Number())
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
			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "first")
			})

			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "second")
			})

			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "third")
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
			h.CycleEnd(func(ctx *fmesh.CycleContext) {
				// Access both FMesh and Cycle through context
				meshName = ctx.FMesh.Name()
				cycleNumber = ctx.Cycle.Number()
				activationCount = ctx.Cycle.ActivationResults().Len()
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
			h.CycleBegin(func(ctx *fmesh.CycleContext) {
				beginCount++
			})
			h.CycleEnd(func(ctx *fmesh.CycleContext) {
				endCount++
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
	var beforeRanFired bool
	var afterRunFired bool

	// Create mesh with chainable error (simulating Run() returning error)
	fm := fmesh.New("test-mesh").
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) {
				beforeRanFired = true
			})
			h.AfterRun(func(fm *fmesh.FMesh) {
				afterRunFired = true
			})
		}).
		WithChainableErr(assert.AnError) // Force error

	_, err := fm.Run()
	require.Error(t, err)

	// AfterRun still fires even on error (like defer)
	assert.True(t, beforeRanFired)
	assert.True(t, afterRunFired)
}

func TestHooks_EmptyMesh(t *testing.T) {
	var beforeRunFired bool
	var afterRunFired bool

	// Create mesh with no components - this will error
	fm := fmesh.New("empty-mesh").
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) {
				beforeRunFired = true
			})
			h.AfterRun(func(fm *fmesh.FMesh) {
				afterRunFired = true
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
			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "first-setup")
			})
		}).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "second-setup")
			})
		}).
		SetupHooks(func(h *fmesh.Hooks) {
			h.BeforeRun(func(fm *fmesh.FMesh) {
				log = append(log, "third-setup")
			})
		})

	fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1))

	_, err := fm.Run()
	require.NoError(t, err)

	// All hooks from all SetupHooks calls should execute in order
	assert.Equal(t, []string{"first-setup", "second-setup", "third-setup"}, log)
}
