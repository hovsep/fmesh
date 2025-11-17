package componenthooks

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentHooks_AllTypes(t *testing.T) {
	var executionLog []string

	c := component.New("processor").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
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
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(42))
			return nil
		})

	c.InputByName("in").PutSignals(signal.New(1))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.Equal(t, component.ActivationCodeOK, result.Code())
	assert.Equal(t, []string{"before", "success", "after"}, executionLog)
}

func TestComponentHooks_OnError(t *testing.T) {
	var errorCaught bool
	var afterFired bool
	testErr := errors.New("test error")

	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
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
		}).
		WithActivationFunc(func(c *component.Component) error {
			return testErr
		})

	c.InputByName("in").PutSignals(signal.New(1))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.True(t, result.IsError())
	assert.True(t, errorCaught, "OnError hook should fire")
	assert.True(t, afterFired, "AfterActivation hook should fire")
}

func TestComponentHooks_OnPanic(t *testing.T) {
	var panicCaught bool
	var afterFired bool

	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
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
		}).
		WithActivationFunc(func(c *component.Component) error {
			panic("oh no!")
		})

	c.InputByName("in").PutSignals(signal.New(1))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	require.True(t, result.IsPanic())
	assert.True(t, panicCaught, "OnPanic hook should fire")
	assert.True(t, afterFired, "AfterActivation hook should fire even after panic")
}

func TestComponentHooks_OnWaitingForInputs(t *testing.T) {
	var waitingCaught bool

	c := component.New("processor").
		AddInputs("data", "config").
		SetupHooks(func(h *component.Hooks) {
			h.OnWaitingForInputs(func(ctx *component.ActivationContext) error {
				waitingCaught = true
				assert.Equal(t, component.ActivationCodeWaitingForInputsClear, ctx.Result.Code())
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			// Wait for config input
			if !c.InputByName("config").Signals().AnyMatch(func(s *signal.Signal) bool { return true }) {
				return component.NewErrWaitForInputs(false)
			}
			return nil
		})

	// Only provide data input, not config
	c.InputByName("data").PutSignals(signal.New(1))
	result := c.MaybeActivate()

	require.True(t, result.Activated())
	assert.True(t, waitingCaught, "OnWaitingForInputs hook should fire")
}

func TestComponentHooks_MultipleHooksPerType(t *testing.T) {
	var log []string

	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
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
		}).
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	c.InputByName("in").PutSignals(signal.New(1))
	c.MaybeActivate()

	assert.Equal(t, []string{"before1", "before2", "success1", "success2"}, log)
}

func TestComponentHooks_NoHooksOnNoInput(t *testing.T) {
	var beforeFired bool

	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				beforeFired = true
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	// No input signals provided
	result := c.MaybeActivate()

	require.False(t, result.Activated())
	require.Equal(t, component.ActivationCodeNoInput, result.Code())
	assert.False(t, beforeFired, "Hooks should not fire when component doesn't activate")
}

func TestComponentHooks_NoHooksOnNoFunction(t *testing.T) {
	var beforeFired bool

	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				beforeFired = true
				return nil
			})
		})
	// No activation function

	c.InputByName("in").PutSignals(signal.New(1))
	result := c.MaybeActivate()

	require.False(t, result.Activated())
	require.Equal(t, component.ActivationCodeNoFunction, result.Code())
	assert.False(t, beforeFired, "Hooks should not fire when component has no activation function")
}

func TestComponentHooks_ContextAccess(t *testing.T) {
	var componentName string
	var activationCode component.ActivationResultCode

	c := component.New("test-component").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
			h.AfterActivation(func(ctx *component.ActivationContext) error {
				componentName = ctx.Component.Name()
				activationCode = ctx.Result.Code()
				assert.Equal(t, 1, ctx.Component.Outputs().ByName("out").Signals().Len())
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(100))
			return nil
		})

	c.InputByName("in").PutSignals(signal.New(1))
	c.MaybeActivate()

	assert.Equal(t, "test-component", componentName)
	assert.Equal(t, component.ActivationCodeOK, activationCode)
}

func TestComponentHooks_IntegrationWithFMesh(t *testing.T) {
	var componentActivations []string

	c1 := component.New("c1").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				componentActivations = append(componentActivations, c.Name())
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(1))
			return nil
		})

	c2 := component.New("c2").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				componentActivations = append(componentActivations, c.Name())
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	c1.OutputByName("out").PipeTo(c2.InputByName("in"))

	fm := fmesh.New("test").AddComponents(c1, c2)
	c1.InputByName("in").PutSignals(signal.New(0))

	_, err := fm.Run()
	require.NoError(t, err)

	// Both components should have activated
	assert.Contains(t, componentActivations, "c1")
	assert.Contains(t, componentActivations, "c2")
}

func TestComponentHooks_ExecutionOrderAcrossComponents(t *testing.T) {
	var executionLog []string

	// Create three components with hooks
	c1 := component.New("c1").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				executionLog = append(executionLog, "c1:before")
				return nil
			})
			h.OnSuccess(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c1:success")
				return nil
			})
			h.AfterActivation(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c1:after")
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(1))
			return nil
		})

	c2 := component.New("c2").
		AddInputs("in").
		AddOutputs("out").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				executionLog = append(executionLog, "c2:before")
				return nil
			})
			h.OnSuccess(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c2:success")
				return nil
			})
			h.AfterActivation(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c2:after")
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(2))
			return nil
		})

	c3 := component.New("c3").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				executionLog = append(executionLog, "c3:before")
				return nil
			})
			h.OnSuccess(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c3:success")
				return nil
			})
			h.AfterActivation(func(ctx *component.ActivationContext) error {
				executionLog = append(executionLog, "c3:after")
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	// Wire: c1, c2 -> c3 (both feed into c3)
	c1.OutputByName("out").PipeTo(c3.InputByName("in"))
	c2.OutputByName("out").PipeTo(c3.InputByName("in"))

	fm := fmesh.New("test").AddComponents(c1, c2, c3)
	c1.InputByName("in").PutSignals(signal.New(0))
	c2.InputByName("in").PutSignals(signal.New(0))

	_, err := fm.Run()
	require.NoError(t, err)

	// Verify all components activated with proper hook order
	// c1 and c2 fire in cycle 1, c3 in cycle 2
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
	c := component.New("processor").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				log = append(log, "setup1")
				return nil
			})
		}).
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				log = append(log, "setup2")
				return nil
			})
		}).
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) error {
				log = append(log, "setup3")
				return nil
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			return nil
		})

	c.InputByName("in").PutSignals(signal.New(1))
	c.MaybeActivate()

	// All hooks from all SetupHooks calls should execute in order
	assert.Equal(t, []string{"setup1", "setup2", "setup3"}, log)
}

func BenchmarkComponentHooks_Overhead(b *testing.B) {
	// Measure overhead of hooks vs no hooks
	b.Run("WithoutHooks", func(b *testing.B) {
		c := component.New("processor").
			AddInputs("in").
			WithActivationFunc(func(c *component.Component) error {
				return nil
			})

		c.InputByName("in").PutSignals(signal.New(1))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.MaybeActivate()
			c.InputByName("in").PutSignals(signal.New(1))
		}
	})

	b.Run("WithHooks", func(b *testing.B) {
		c := component.New("processor").
			AddInputs("in").
			SetupHooks(func(h *component.Hooks) {
				h.BeforeActivation(func(c *component.Component) error { return nil })
				h.OnSuccess(func(ctx *component.ActivationContext) error { return nil })
				h.AfterActivation(func(ctx *component.ActivationContext) error { return nil })
			}).
			WithActivationFunc(func(c *component.Component) error {
				return nil
			})

		c.InputByName("in").PutSignals(signal.New(1))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.MaybeActivate()
			c.InputByName("in").PutSignals(signal.New(1))
		}
	})
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
