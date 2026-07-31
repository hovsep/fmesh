package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

// Component hooks: which ones fire, in what order, and what the activation
// context carries.

// recorder collects hook names in firing order.
type recorder struct {
	events []string
}

func (r *recorder) add(event string) { r.events = append(r.events, event) }

// componentWith builds a single-component mesh whose activation runs af.
func componentWith(t *testing.T, af component.ActivationFunc, hooks func(*component.Hooks)) *component.Component {
	t.Helper()
	c := testutil.MustComponent("processor",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(af))
	if hooks != nil {
		c.SetupHooks(hooks)
	}
	return c
}

func TestComponentHooks_FiringOrderOnSuccess(t *testing.T) {
	var log recorder

	c := componentWith(t,
		func(_ context.Context, this *component.Component) error {
			log.add("activation")
			return this.OutputByName("out").PutSignals(signal.New(1))
		},
		func(h *component.Hooks) {
			h.BeforeActivation(func(context.Context, *component.Component) error {
				log.add("before")
				return nil
			})
			h.OnActivation(func(context.Context, *component.Component) error {
				log.add("onActivation")
				return nil
			})
			h.OnSuccess(func(context.Context, *component.ActivationContext) error {
				log.add("success")
				return nil
			})
			h.OnError(func(context.Context, *component.ActivationContext) error {
				log.add("error")
				return nil
			})
			h.AfterActivation(func(context.Context, *component.ActivationContext) error {
				log.add("after")
				return nil
			})
		})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	require.True(t, c.MaybeActivate(context.Background()).Activated())

	// OnActivation runs after the main function, as part of the same activation.
	assert.Equal(t, []string{"before", "activation", "onActivation", "success", "after"}, log.events)
}

func TestComponentHooks_OutcomeSpecificHooks(t *testing.T) {
	tests := []struct {
		name     string
		activate component.ActivationFunc
		hook     func(*component.Hooks, *recorder)
		wantCode component.ActivationResultCode
		wantLog  []string
	}{
		{
			name:     "error",
			activate: func(context.Context, *component.Component) error { return errors.New("boom") },
			hook: func(h *component.Hooks, log *recorder) {
				h.OnError(func(context.Context, *component.ActivationContext) error {
					log.add("error")
					return nil
				})
			},
			wantCode: component.ActivationCodeReturnedError,
			wantLog:  []string{"error", "after"},
		},
		{
			name:     "panic",
			activate: func(context.Context, *component.Component) error { panic("boom") },
			hook: func(h *component.Hooks, log *recorder) {
				h.OnPanic(func(context.Context, *component.ActivationContext) error {
					log.add("panic")
					return nil
				})
			},
			wantCode: component.ActivationCodePanicked,
			wantLog:  []string{"panic", "after"},
		},
		{
			name:     "waiting, dropping inputs",
			activate: func(context.Context, *component.Component) error { return component.ErrWaitDroppingInputs },
			hook: func(h *component.Hooks, log *recorder) {
				h.OnWaitingForInputs(func(context.Context, *component.ActivationContext) error {
					log.add("waiting")
					return nil
				})
			},
			wantCode: component.ActivationCodeWaitingForInputsClear,
			wantLog:  []string{"waiting", "after"},
		},
		{
			name:     "waiting, keeping inputs",
			activate: func(context.Context, *component.Component) error { return component.ErrWaitKeepingInputs },
			hook: func(h *component.Hooks, log *recorder) {
				h.OnWaitingForInputs(func(context.Context, *component.ActivationContext) error {
					log.add("waiting")
					return nil
				})
			},
			wantCode: component.ActivationCodeWaitingForInputsKeep,
			wantLog:  []string{"waiting", "after"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log recorder
			c := componentWith(t, tt.activate, func(h *component.Hooks) {
				tt.hook(h, &log)
				h.AfterActivation(func(context.Context, *component.ActivationContext) error {
					log.add("after")
					return nil
				})
			})

			require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
			result := c.MaybeActivate(context.Background())

			assert.Equal(t, tt.wantCode, result.Code())
			assert.Equal(t, tt.wantLog, log.events)
		})
	}
}

func TestComponentHooks_ActivationContextCarriesComponentAndResult(t *testing.T) {
	var gotName string
	var gotCode component.ActivationResultCode
	var gotOutputs int

	c := componentWith(t,
		func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New(100))
		},
		func(h *component.Hooks) {
			h.AfterActivation(func(_ context.Context, activation *component.ActivationContext) error {
				gotName = activation.Component.Name()
				gotCode = activation.Result.Code()
				// Outputs are still on the port here — the drain happens later.
				gotOutputs = activation.Component.OutputByName("out").Signals().Len()
				return nil
			})
		})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	c.MaybeActivate(context.Background())

	assert.Equal(t, "processor", gotName)
	assert.Equal(t, component.ActivationCodeOK, gotCode)
	assert.Equal(t, 1, gotOutputs)
}

func TestComponentHooks_NoActivationMeansNoHooks(t *testing.T) {
	var log recorder

	c := componentWith(t,
		func(context.Context, *component.Component) error { return nil },
		func(h *component.Hooks) {
			h.BeforeActivation(func(context.Context, *component.Component) error {
				log.add("before")
				return nil
			})
			h.AfterActivation(func(context.Context, *component.ActivationContext) error {
				log.add("after")
				return nil
			})
		})

	// No input: the component never activates, so nothing fires.
	result := c.MaybeActivate(context.Background())

	require.Equal(t, component.ActivationCodeNoInput, result.Code())
	assert.Empty(t, log.events)
}

func TestComponentHooks_AccumulateInRegistrationOrder(t *testing.T) {
	var log recorder

	// Hooks of the same type, and repeated SetupHooks calls, both accumulate.
	c := componentWith(t, func(context.Context, *component.Component) error { return nil },
		func(h *component.Hooks) {
			h.BeforeActivation(func(context.Context, *component.Component) error {
				log.add("first")
				return nil
			})
			h.BeforeActivation(func(context.Context, *component.Component) error {
				log.add("second")
				return nil
			})
		})
	c.SetupHooks(func(h *component.Hooks) {
		h.BeforeActivation(func(context.Context, *component.Component) error {
			log.add("third")
			return nil
		})
	})

	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))
	c.MaybeActivate(context.Background())

	assert.Equal(t, []string{"first", "second", "third"}, log.events)
}

func TestComponentHooks_FireForEveryComponentInAMeshRun(t *testing.T) {
	var log recorder

	track := func(c *component.Component) *component.Component {
		return c.SetupHooks(func(h *component.Hooks) {
			h.BeforeActivation(func(_ context.Context, this *component.Component) error {
				log.add(this.Name() + ":before")
				return nil
			})
			h.AfterActivation(func(_ context.Context, activation *component.ActivationContext) error {
				log.add(activation.Component.Name() + ":after")
				return nil
			})
		})
	}

	producer := track(testutil.MustComponent("producer",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New(1))
		})))
	consumer := track(testutil.MustComponent("consumer",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })))

	fm := testutil.MustFMesh("hooked")
	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	require.NoError(t, producer.InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	// The producer activates in cycle 1, the consumer in cycle 2 once the signal
	// has been drained to it.
	assert.Equal(t, []string{
		"producer:before", "producer:after",
		"consumer:before", "consumer:after",
	}, log.events)
}

func TestComponentHooks_HookFailureStopsTheMesh(t *testing.T) {
	c := componentWith(t,
		func(context.Context, *component.Component) error { return nil },
		func(h *component.Hooks) {
			h.BeforeActivation(func(context.Context, *component.Component) error {
				return errors.New("hook says no")
			})
		})

	fm := testutil.MustFMesh("failing-hook")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())

	// A hook failure is an activation error, so the default strategy stops the run.
	require.ErrorIs(t, err, fmesh.ErrHitAnErrorOrPanic)
	assert.Contains(t, err.Error(), "hook says no")
}
