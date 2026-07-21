package component

import (
	"bytes"
	"errors"
	"testing"

	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent_WithActivationFunc(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		f := func(this *Component) error {
			if out := this.OutputByName("out1"); out != nil {
				return out.PutSignals(signal.New(23))
			}
			return nil
		}
		c := mustNew("c1", WithOutputs("out1"), WithActivationFunc(f))
		assert.NotNil(t, c.f)

		// Verify the assigned function produces the same output as the original
		dummy1 := mustNew("d1", WithOutputs("out1"))
		dummy2 := mustNew("d2", WithOutputs("out1"))
		err1 := c.f(dummy1)
		err2 := f(dummy2)
		assert.Equal(t, err1, err2)
		assert.ElementsMatch(t, dummy1.OutputByName("out1").Signals().All(), dummy2.OutputByName("out1").Signals().All())
	})

	t.Run("WithActivationFunc replaces previous value", func(t *testing.T) {
		first := func(this *Component) error { return nil }
		second := func(this *Component) error { return errors.New("second") }
		c := mustNew("c1", WithActivationFunc(first), WithActivationFunc(second))
		require.NotNil(t, c.f)
		err := c.f(c)
		assert.EqualError(t, err, "second")
	})
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
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}),
				)
				require.NoError(t, err)
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				SetActivationCode(ActivationCodeNoInput),
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithActivationFunc(func(this *Component) error {
						return errors.New("test error")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				SetActivationCode(ActivationCodeReturnedError).
				AddActivationError(errors.New("component returned an error: test error")),
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				SetActivationCode(ActivationCodeOK),
		},
		{
			name: "component panicked with error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						panic(errors.New("oh shrimps"))
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				SetActivationCode(ActivationCodePanicked).
				AddActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						panic("oh shrimps")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				SetActivationCode(ActivationCodePanicked).
				AddActivationError(errors.New("panicked with: oh shrimps")),
		},
		{
			name: "component is waiting for inputs",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1", "i2"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return ErrWaitingForInputs
						}
						return nil
					}),
				)
				require.NoError(t, err)
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
				c, err := New("c1",
					WithInputs("i1", "i2"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return ErrWaitingForInputsKeep
						}
						return nil
					}),
				)
				require.NoError(t, err)
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
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(this *Component) error {
						this.Logger().Println("This must not be logged, as component must not activate")
						return nil
					}),
				)
				require.NoError(t, err)
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(false).
				SetActivationCode(ActivationCodeNoInput),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Empty(t, output)
			},
		},
		{
			name: "activated with error, with logging",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithActivationFunc(func(this *Component) error {
						this.Logger().Println("This line must be logged")
						return errors.New("test error")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1").
				SetActivated(true).
				SetActivationCode(ActivationCodeReturnedError).
				AddActivationError(errors.New("component returned an error: test error")),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.NotEmpty(t, output)
				assert.Contains(t, string(output), "c1: This line must be logged")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loggerOutput bytes.Buffer

			component := tt.getComponent()
			component.Logger().SetOutput(&loggerOutput)

			gotActivationResult := component.MaybeActivate()
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
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(this *Component) error { return nil }),
		)
		require.NoError(t, err)
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
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(this *Component) error { return nil }),
		)
		require.NoError(t, err)
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
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(this *Component) error {
				return errors.New("component error")
			}),
		)
		require.NoError(t, err)
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
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(this *Component) error { return nil }),
		)
		require.NoError(t, err)
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
