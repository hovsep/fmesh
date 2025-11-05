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
			h.BeforeActivation(func(c *component.Component) {
				executionLog = append(executionLog, "before")
			})

			h.OnSuccess(func(ctx *component.ActivationContext) {
				executionLog = append(executionLog, "success")
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				executionLog = append(executionLog, "after")
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
			h.OnError(func(ctx *component.ActivationContext) {
				errorCaught = true
				assert.Equal(t, component.ActivationCodeReturnedError, ctx.Result.Code())
				assert.Error(t, ctx.Result.ActivationError())
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				afterFired = true
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
			h.OnPanic(func(ctx *component.ActivationContext) {
				panicCaught = true
				assert.Equal(t, component.ActivationCodePanicked, ctx.Result.Code())
				assert.Error(t, ctx.Result.ActivationError())
			})

			h.AfterActivation(func(ctx *component.ActivationContext) {
				afterFired = true
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
			h.OnWaitingForInputs(func(ctx *component.ActivationContext) {
				waitingCaught = true
				assert.Equal(t, component.ActivationCodeWaitingForInputsClear, ctx.Result.Code())
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
			h.BeforeActivation(func(c *component.Component) {
				log = append(log, "before1")
			})
			h.BeforeActivation(func(c *component.Component) {
				log = append(log, "before2")
			})
			h.OnSuccess(func(ctx *component.ActivationContext) {
				log = append(log, "success1")
			})
			h.OnSuccess(func(ctx *component.ActivationContext) {
				log = append(log, "success2")
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
			h.BeforeActivation(func(c *component.Component) {
				beforeFired = true
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
			h.BeforeActivation(func(c *component.Component) {
				beforeFired = true
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
			h.AfterActivation(func(ctx *component.ActivationContext) {
				componentName = ctx.Component.Name()
				activationCode = ctx.Result.Code()
				assert.Equal(t, 1, ctx.Component.Outputs().ByName("out").Signals().Len())
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
			h.BeforeActivation(func(c *component.Component) {
				componentActivations = append(componentActivations, c.Name())
			})
		}).
		WithActivationFunc(func(c *component.Component) error {
			c.OutputByName("out").PutSignals(signal.New(1))
			return nil
		})

	c2 := component.New("c2").
		AddInputs("in").
		SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(c *component.Component) {
				componentActivations = append(componentActivations, c.Name())
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
