package component

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent_WithActivationFunc(t *testing.T) {
	type args struct {
		f ActivationFunc
	}
	tests := []struct {
		name      string
		component *Component
		args      args
	}{
		{
			name: "happy path",
			component: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddOutputs("out1"))
				return c
			}(),
			args: args{
				f: func(this *Component) error {
					if out := this.OutputByName("out1"); out != nil {
						return out.PutSignals(signal.New(23))
					}
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithActivationFunc(tt.args.f)

			// Compare activation functions by they result and error
			dummyComponent1, err := New("c1")
			require.NoError(t, err)
			require.NoError(t, dummyComponent1.AddInputs("i1", "i2"))
			require.NoError(t, dummyComponent1.AddOutputs("o1", "o2"))

			dummyComponent2, err := New("c2")
			require.NoError(t, err)
			require.NoError(t, dummyComponent2.AddInputs("i1", "i2"))
			require.NoError(t, dummyComponent2.AddOutputs("o1", "o2"))

			err1 := componentAfter.f(dummyComponent1)
			err2 := tt.args.f(dummyComponent2)
			assert.Equal(t, err1, err2)

			// Compare signals without keys (because they are random)
			o1Signals1, _ := dummyComponent1.OutputByName("o1").Signals().All()
			o1Signals2, _ := dummyComponent2.OutputByName("o1").Signals().All()
			assert.ElementsMatch(t, o1Signals1, o1Signals2)

			o2Signals1, _ := dummyComponent1.OutputByName("o2").Signals().All()
			o2Signals2, _ := dummyComponent2.OutputByName("o2").Signals().All()
			assert.ElementsMatch(t, o2Signals1, o2Signals2)
		})
	}
}

func TestComponent_MaybeActivate(t *testing.T) {
	tests := []struct {
		name                 string
		getComponent         func() *Component
		wantActivationResult *ActivationResult
		loggerAssertions     func(t *testing.T, output []byte)
	}{
		{
			name: "component with activation func, but no inputs",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
				})
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoInput),
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				c.WithActivationFunc(func(this *Component) error {
					return errors.New("test error")
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithActivationError(errors.New("component returned an error: test error")),
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeOK),
		},
		{
			name: "component panicked with error",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					panic(errors.New("oh shrimps"))
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					panic("oh shrimps")
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodePanicked).
				WithActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component is waiting for inputs",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1", "i2"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
						return ErrWaitingForInputs
					}
					return nil
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: &ActivationResult{
				componentName:    "c1",
				activated:        true,
				code:             ActivationCodeWaitingForInputsClear,
				activationErrors: []error{ErrWaitingForInputs},
			},
		},
		{
			name: "component is waiting for inputs and wants to keep them",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1", "i2"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
						return ErrWaitingForInputsKeep
					}
					return nil
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: &ActivationResult{
				componentName:    "c1",
				activated:        true,
				code:             ActivationCodeWaitingForInputsKeep,
				activationErrors: []error{ErrWaitingForInputsKeep},
			},
		},
		{
			name: "component not activated, logger must be empty",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				require.NoError(t, c.AddOutputs("o1"))
				c.WithActivationFunc(func(this *Component) error {
					this.Logger().Println("This must not be logged, as component must not activate")
					return nil
				})
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				WithActivationCode(ActivationCodeNoInput),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Empty(t, output)
			},
		},
		{
			name: "activated with error, with logging",
			getComponent: func() *Component {
				c, err := New("c1")
				require.NoError(t, err)
				require.NoError(t, c.AddInputs("i1"))
				c.WithActivationFunc(func(this *Component) error {
					this.logger.Println("This line must be logged")
					return errors.New("test error")
				})
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				WithActivationCode(ActivationCodeReturnedError).
				WithActivationError(errors.New("component returned an error: test error")),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Len(t, output, 2+3+21+24) // lengths of component name, prefix, flags and logged message
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.Default()

			var loggerOutput bytes.Buffer
			logger.SetOutput(&loggerOutput)

			gotActivationResult := tt.getComponent().WithLogger(logger).MaybeActivate()
			assert.Equal(t, tt.wantActivationResult.Activated(), gotActivationResult.Activated())
			assert.Equal(t, tt.wantActivationResult.ComponentName(), gotActivationResult.ComponentName())
			assert.Equal(t, tt.wantActivationResult.Code(), gotActivationResult.Code())
			if tt.wantActivationResult.IsError() {
				assert.EqualError(t, gotActivationResult.ActivationError(), tt.wantActivationResult.ActivationError().Error())
			} else {
				assert.False(t, gotActivationResult.IsError())
			}

			if tt.loggerAssertions != nil {
				tt.loggerAssertions(t, loggerOutput.Bytes())
			}
		})
	}
}

func TestComponent_MaybeActivate_HookFailures(t *testing.T) {
	t.Run("beforeActivation hook fails: ActivationCodeHookFailed, not activated, error captured", func(t *testing.T) {
		c, err := New("c1")
		require.NoError(t, err)
		require.NoError(t, c.AddInputs("i1"))
		c.WithActivationFunc(func(this *Component) error { return nil })
		c.SetupHooks(func(h *Hooks) {
			h.BeforeActivation(func(_ *Component) error {
				return errors.New("before hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate()

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		assert.False(t, result.Activated())
		require.Error(t, result.ActivationError())
		require.ErrorContains(t, result.ActivationError(), "before hook error")
		assert.Len(t, result.ActivationErrors(), 1)
	})

	t.Run("onSuccess hook fails: ActivationCodeHookFailed, error accumulated", func(t *testing.T) {
		c, err := New("c1")
		require.NoError(t, err)
		require.NoError(t, c.AddInputs("i1"))
		c.WithActivationFunc(func(this *Component) error { return nil })
		c.SetupHooks(func(h *Hooks) {
			h.OnSuccess(func(_ *ActivationContext) error {
				return errors.New("onSuccess hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate()

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		require.Error(t, result.ActivationError())
		require.ErrorContains(t, result.ActivationError(), "onSuccess hook error")
		assert.Len(t, result.ActivationErrors(), 1)
	})

	t.Run("onError hook fails: ActivationCodeHookFailed, both errors accumulated", func(t *testing.T) {
		c, err := New("c1")
		require.NoError(t, err)
		require.NoError(t, c.AddInputs("i1"))
		c.WithActivationFunc(func(this *Component) error {
			return errors.New("component error")
		})
		c.SetupHooks(func(h *Hooks) {
			h.OnError(func(_ *ActivationContext) error {
				return errors.New("onError hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate()

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		assert.Len(t, result.ActivationErrors(), 2)
		require.ErrorContains(t, result.ActivationError(), "component error")
		assert.ErrorContains(t, result.ActivationError(), "onError hook error")
	})

	t.Run("afterActivation hook fails: ActivationCodeHookFailed, error accumulated", func(t *testing.T) {
		c, err := New("c1")
		require.NoError(t, err)
		require.NoError(t, c.AddInputs("i1"))
		c.WithActivationFunc(func(this *Component) error { return nil })
		c.SetupHooks(func(h *Hooks) {
			h.AfterActivation(func(_ *ActivationContext) error {
				return errors.New("afterActivation hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate()

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		require.Error(t, result.ActivationError())
		require.ErrorContains(t, result.ActivationError(), "afterActivation hook error")
		assert.Len(t, result.ActivationErrors(), 1)
	})
}
