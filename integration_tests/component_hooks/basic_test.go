package componenthooks

import (
	"errors"
	"sync"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustFMesh(name string, opts ...fmesh.Option) *fmesh.FMesh {
	fm, err := fmesh.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

func TestComponentHooks_AllTypes(t *testing.T) {
	var executionLog []string

	c := mustComponent("processor",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			executionLog = append(executionLog, "before")
			return nil
		})

		h.OnSuccess(func(ctx *component.ActivationContext) error {
			executionLog = append(executionLog, "success")
			return nil
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			executionLog = append(executionLog, "after")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return c.OutputByName("out").PutSignals(signal.New(42))
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.Equal(t, component.ActivationCodeOK, result.Code())
	assert.Equal(t, []string{"before", "success", "after"}, executionLog)
}

func TestComponentHooks_OnError(t *testing.T) {
	var errorCaught bool
	var afterFired bool
	testErr := errors.New("test error")

	c := mustComponent("processor",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnError(func(ctx *component.ActivationContext) error {
			errorCaught = true
			assert.Equal(t, component.ActivationCodeReturnedError, ctx.Result.Code())
			require.Error(t, ctx.Result.ActivationError())
			return nil
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			afterFired = true
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return testErr
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.True(t, result.IsError())
	assert.True(t, errorCaught, "OnError hook should fire")
	assert.True(t, afterFired, "AfterActivation hook should fire")
}

func TestComponentHooks_OnPanic(t *testing.T) {
	var panicCaught bool
	var afterFired bool

	c := mustComponent("processor",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnPanic(func(ctx *component.ActivationContext) error {
			panicCaught = true
			assert.Equal(t, component.ActivationCodePanicked, ctx.Result.Code())
			require.Error(t, ctx.Result.ActivationError())
			return nil
		})

		h.AfterActivation(func(ctx *component.ActivationContext) error {
			afterFired = true
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		panic("oh no!")
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.True(t, result.IsPanic())
	assert.True(t, panicCaught, "OnPanic hook should fire")
	assert.True(t, afterFired, "AfterActivation hook should fire even after panic")
}

func TestComponentHooks_OnWaitingForInputs(t *testing.T) {
	var waitingCaught bool

	c := mustComponent("processor",
		component.WithInputs("data", "config"),
	).SetupHooks(func(h *component.Hooks) {
		h.OnWaitingForInputs(func(ctx *component.ActivationContext) error {
			waitingCaught = true
			assert.Equal(t, component.ActivationCodeWaitingForInputsClear, ctx.Result.Code())
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		// Wait for config input
		if !c.InputByName("config").Signals().Any(func(s *signal.Signal) bool { return true }) {
			return component.ErrWaitingForInputs
		}
		return nil
	})

	// Only provide data input, not config
	require.NoError(t, c.InputByName("data").PutSignals(signal.New(1)))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	assert.True(t, waitingCaught, "OnWaitingForInputs hook should fire")
}

func TestComponentHooks_MultipleHooksPerType(t *testing.T) {
	var log []string

	c := mustComponent("processor",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log = append(log, "before1")
			return nil
		})
		h.BeforeActivation(func(c *component.Component) error {
			log = append(log, "before2")
			return nil
		})
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			log = append(log, "success1")
			return nil
		})
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			log = append(log, "success2")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return nil
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	c.MaybeActivate()

	assert.Equal(t, []string{"before1", "before2", "success1", "success2"}, log)
}

func TestComponentHooks_NoHooksOnNoInput(t *testing.T) {
	var beforeFired bool

	c := mustComponent("processor",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			beforeFired = true
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return nil
	})

	// No input signals provided
	result := c.MaybeActivate()

	require.False(t, result.Activated())
	require.Equal(t, component.ActivationCodeNoInput, result.Code())
	assert.False(t, beforeFired, "Hooks should not fire when component doesn't activate")
}

func TestComponentHooks_ContextAccess(t *testing.T) {
	var componentName string
	var activationCode component.ActivationResultCode

	c := mustComponent("test-component",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.AfterActivation(func(ctx *component.ActivationContext) error {
			componentName = ctx.Component.Name()
			activationCode = ctx.Result.Code()
			assert.Equal(t, 1, ctx.Component.Outputs().ByName("out").Signals().Len())
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return c.OutputByName("out").PutSignals(signal.New(100))
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	c.MaybeActivate()

	assert.Equal(t, "test-component", componentName)
	assert.Equal(t, component.ActivationCodeOK, activationCode)
}

func TestComponentHooks_IntegrationWithFMesh(t *testing.T) {
	var log hookLog

	c1 := mustComponent("c1",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log.add(c.Name())
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return c.OutputByName("out").PutSignals(signal.New(1))
	})

	c2 := mustComponent("c2",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log.add(c.Name())
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return nil
	})

	require.NoError(t, c1.OutputByName("out").PipeTo(c2.InputByName("in")))

	fm := mustFMesh("test")
	require.NoError(t, fm.AddComponents(c1, c2))
	require.NoError(t, c1.InputByName("in").PutSignals(signal.New(0)))

	_, err := fm.Run()
	require.NoError(t, err)

	// Both components should have activated
	componentActivations := log.snapshot()
	assert.Contains(t, componentActivations, "c1")
	assert.Contains(t, componentActivations, "c2")
}

func TestComponentHooks_ExecutionOrderAcrossComponents(t *testing.T) {
	var log hookLog

	// Create three components with hooks
	c1 := mustComponent("c1",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log.add("c1:before")
			return nil
		})
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			log.add("c1:success")
			return nil
		})
		h.AfterActivation(func(ctx *component.ActivationContext) error {
			log.add("c1:after")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return c.OutputByName("out").PutSignals(signal.New(1))
	})

	c2 := mustComponent("c2",
		component.WithInputs("in"),
		component.WithOutputs("out"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log.add("c2:before")
			return nil
		})
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			log.add("c2:success")
			return nil
		})
		h.AfterActivation(func(ctx *component.ActivationContext) error {
			log.add("c2:after")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return c.OutputByName("out").PutSignals(signal.New(2))
	})

	c3 := mustComponent("c3",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log.add("c3:before")
			return nil
		})
		h.OnSuccess(func(ctx *component.ActivationContext) error {
			log.add("c3:success")
			return nil
		})
		h.AfterActivation(func(ctx *component.ActivationContext) error {
			log.add("c3:after")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return nil
	})

	// Wire: c1, c2 -> c3 (both feed into c3)
	require.NoError(t, c1.OutputByName("out").PipeTo(c3.InputByName("in")))
	require.NoError(t, c2.OutputByName("out").PipeTo(c3.InputByName("in")))

	fm := mustFMesh("test")
	require.NoError(t, fm.AddComponents(c1, c2, c3))
	require.NoError(t, c1.InputByName("in").PutSignals(signal.New(0)))
	require.NoError(t, c2.InputByName("in").PutSignals(signal.New(0)))

	_, err := fm.Run()
	require.NoError(t, err)

	executionLog := log.snapshot()

	// Verify all components activated with proper hook order
	assert.Contains(t, executionLog, "c1:before")
	assert.Contains(t, executionLog, "c1:success")
	assert.Contains(t, executionLog, "c1:after")
	assert.Contains(t, executionLog, "c2:before")
	assert.Contains(t, executionLog, "c2:success")
	assert.Contains(t, executionLog, "c2:after")
	assert.Contains(t, executionLog, "c3:before")
	assert.Contains(t, executionLog, "c3:success")
	assert.Contains(t, executionLog, "c3:after")

	// Each component's hooks execute in order: before -> success -> after
	c1BeforeIdx := indexOf(executionLog, "c1:before")
	c1SuccessIdx := indexOf(executionLog, "c1:success")
	c1AfterIdx := indexOf(executionLog, "c1:after")
	assert.True(t, c1BeforeIdx < c1SuccessIdx && c1SuccessIdx < c1AfterIdx)

	c2BeforeIdx := indexOf(executionLog, "c2:before")
	c2SuccessIdx := indexOf(executionLog, "c2:success")
	c2AfterIdx := indexOf(executionLog, "c2:after")
	assert.True(t, c2BeforeIdx < c2SuccessIdx && c2SuccessIdx < c2AfterIdx)
}

func TestComponentHooks_MultipleSetupCalls(t *testing.T) {
	var log []string

	// Multiple SetupHooks calls should accumulate hooks
	c := mustComponent("processor",
		component.WithInputs("in"),
	).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log = append(log, "setup1")
			return nil
		})
	}).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log = append(log, "setup2")
			return nil
		})
	}).SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(c *component.Component) error {
			log = append(log, "setup3")
			return nil
		})
	}).WithActivationFunc(func(c *component.Component) error {
		return nil
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	c.MaybeActivate()

	// All hooks from all SetupHooks calls should execute in order
	assert.Equal(t, []string{"setup1", "setup2", "setup3"}, log)
}

func BenchmarkComponentHooks_Overhead(b *testing.B) {
	// Measure overhead of hooks vs no hooks
	b.Run("WithoutHooks", func(b *testing.B) {
		c := mustComponent("processor",
			component.WithInputs("in"),
			component.WithActivationFunc(func(c *component.Component) error {
				return nil
			}),
		)

		if err := c.InputByName("in").PutSignals(signal.New(1)); err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for range b.N {
			c.MaybeActivate()
			if err := c.InputByName("in").PutSignals(signal.New(1)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WithHooks", func(b *testing.B) {
		c := mustComponent("processor",
			component.WithInputs("in"),
			component.WithActivationFunc(func(c *component.Component) error {
				return nil
			}),
		).SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error { return nil })
			h.OnSuccess(func(ctx *component.ActivationContext) error { return nil })
			h.AfterActivation(func(ctx *component.ActivationContext) error { return nil })
		})

		if err := c.InputByName("in").PutSignals(signal.New(1)); err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for range b.N {
			c.MaybeActivate()
			if err := c.InputByName("in").PutSignals(signal.New(1)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// hookLog collects hook events from parallel component activations safely.
type hookLog struct {
	mu    sync.Mutex
	lines []string
}

func (l *hookLog) add(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, s)
}

func (l *hookLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.lines))
	copy(out, l.lines)
	return out
}

// Helper function for finding index in slice.
func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
