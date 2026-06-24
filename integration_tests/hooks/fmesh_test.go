package hooks

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustFMesh(name string, opts ...fmesh.Option) *fmesh.FMesh {
	fm, err := fmesh.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func TestHooks_AllTypes(t *testing.T) {
	// Track hook execution
	var executionLog []string

	// Create a simple component
	c := mustComponent("processor",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	// Create a mesh with all hook types.
	// Hooks are registered before AddComponents so OnComponentAdded fires.
	fm := mustFMesh("test-mesh")
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.OnComponentAdded(func(ctx *fmesh.ComponentAddedContext) error {
			executionLog = append(executionLog, "componentAdded")
			return nil
		})

		h.BeforeRun(func(fm *fmesh.FMesh) error {
			executionLog = append(executionLog, "beforeRun")
			return nil
		})

		h.AfterRun(func(fm *fmesh.FMesh) error {
			executionLog = append(executionLog, "afterRun")
			return nil
		})

		h.BeforeCycle(func(ctx *fmesh.CycleContext) error {
			executionLog = append(executionLog, "beforeCycle")
			return nil
		})

		h.AfterCycle(func(ctx *fmesh.CycleContext) error {
			executionLog = append(executionLog, "afterCycle")
			return nil
		})
	})
	require.NoError(t, fm.AddComponents(c))

	// Add initial input
	require.NoError(t, fm.ComponentByName("processor").InputByName("in").PutSignals(signal.New(1)))

	// Run mesh
	_, err := fm.Run()
	require.NoError(t, err)

	// Verify the exact execution order: componentAdded -> beforeRun -> cycles -> afterRun
	// OnComponentAdded fires during AddComponents, before BeforeRun.
	// Cycle hooks fire twice: once for processing, once for completion
	assert.Equal(t, []string{
		"componentAdded",
		"beforeRun",
		"beforeCycle",
		"afterCycle",
		"beforeCycle",
		"afterCycle",
		"afterRun",
	}, executionLog)
}

func TestHooks_CycleContext(t *testing.T) {
	cycleNumbers := []int{}

	// Create a simple component
	c := mustComponent("processor",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	fm := mustFMesh("test-mesh")
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeCycle(func(ctx *fmesh.CycleContext) error {
			cycleNumbers = append(cycleNumbers, ctx.Cycle.Number())
			return nil
		})
	})

	// Add input for multiple cycles
	require.NoError(t, fm.ComponentByName("processor").InputByName("in").
		PutSignals(signal.New(1), signal.New(2), signal.New(3)))

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

	c := mustComponent("test",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	fm := mustFMesh("test-mesh")
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
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

	require.NoError(t, fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run()
	require.NoError(t, err)

	// Verify all hooks executed in order
	assert.Equal(t, []string{"first", "second", "third"}, log)
}

func TestHooks_ContextAccess(t *testing.T) {
	var meshName string
	var cycleNumber int
	var activationCount int

	c := mustComponent("test",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	fm := mustFMesh("my-mesh")
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.AfterCycle(func(ctx *fmesh.CycleContext) error {
			// Access both FMesh and Cycle through context
			meshName = ctx.FMesh.Name()
			cycleNumber = ctx.Cycle.Number()
			activationCount = ctx.Cycle.ActivationResults().Len()
			return nil
		})
	})

	require.NoError(t, fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1)))

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

	c := mustComponent("processor",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	fm := mustFMesh("test-mesh")
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.BeforeCycle(func(ctx *fmesh.CycleContext) error {
			beginCount++
			return nil
		})
		h.AfterCycle(func(ctx *fmesh.CycleContext) error {
			endCount++
			return nil
		})
	})

	// Add multiple signals to trigger multiple cycles
	require.NoError(t, fm.ComponentByName("processor").InputByName("in").
		PutSignals(signal.New(1), signal.New(2), signal.New(3)))

	runtimeInfo, err := fm.Run()
	require.NoError(t, err)

	// Verify hooks fire exactly once per cycle
	actualCycles := runtimeInfo.Cycles.Len()
	assert.Equal(t, actualCycles, beginCount, "BeforeCycle should fire once per cycle")
	assert.Equal(t, actualCycles, endCount, "AfterCycle should fire once per cycle")
	assert.Equal(t, beginCount, endCount, "Begin and End should fire same number of times")
}

func TestHooks_RunWithError(t *testing.T) {
	var beforeRunFired bool
	var afterRunFired bool

	// Create a mesh that will have no components (simulating a broken mesh)
	fm := mustFMesh("test-mesh")
	fm.SetupHooks(func(h *fmesh.Hooks) {
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
	require.Error(t, err)

	// Empty mesh returns error; hooks may or may not fire depending on implementation
	_ = beforeRunFired
	_ = afterRunFired
}

func TestHooks_EmptyMesh(t *testing.T) {
	var beforeRunFired bool
	var afterRunFired bool

	// Create mesh with no components - this will error
	fm := mustFMesh("empty-mesh")
	fm.SetupHooks(func(h *fmesh.Hooks) {
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

	c := mustComponent("test",
		component.WithInputs("in"),
		component.WithActivationFunc(func(c *component.Component) error {
			return nil
		}),
	)

	// Multiple SetupHooks calls should accumulate hooks
	fm := mustFMesh("test-mesh")
	require.NoError(t, fm.AddComponents(c))
	fm.SetupHooks(func(h *fmesh.Hooks) {
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

	require.NoError(t, fm.ComponentByName("test").InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run()
	require.NoError(t, err)

	// All hooks from all SetupHooks calls should execute in order
	assert.Equal(t, []string{"first-setup", "second-setup", "third-setup"}, log)
}
