package component

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent_WithActivationFunc(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		f := func(_ context.Context, this *Component) error {
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
		err1 := c.f(context.Background(), dummy1)
		err2 := f(context.Background(), dummy2)
		assert.Equal(t, err1, err2)
		assert.ElementsMatch(t, dummy1.OutputByName("out1").Signals().All(), dummy2.OutputByName("out1").Signals().All())
	})

	t.Run("WithActivationFunc replaces previous value", func(t *testing.T) {
		first := func(_ context.Context, this *Component) error { return nil }
		second := func(_ context.Context, this *Component) error { return errors.New("second") }
		c := mustNew("c1", WithActivationFunc(first), WithActivationFunc(second))
		require.NotNil(t, c.f)
		err := c.f(context.Background(), c)
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
					WithActivationFunc(func(_ context.Context, this *Component) error {
						return port.ForwardSignals(context.Background(), this.InputByName("i1"), this.OutputByName("o1"))
					}),
				)
				require.NoError(t, err)
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodeNoInput),
		},
		{
			name: "activated with error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						return errors.New("test error")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodeReturnedError, errors.New("component returned an error: test error")),
		},
		{
			name: "activated without error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						return port.ForwardSignals(context.Background(), this.InputByName("i1"), this.OutputByName("o1"))
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodeOK),
		},
		{
			name: "component panicked with error",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						panic(errors.New("oh shrimps"))
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodePanicked, errors.New("panicked: oh shrimps")),
		},
		{
			name: "component panicked with string",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						panic("oh shrimps")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodePanicked, errors.New("panicked: oh shrimps")),
		},
		{
			name: "component is waiting for inputs",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1", "i2"),
					WithOutputs("o1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return ErrWaitDroppingInputs
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
					WithActivationFunc(func(_ context.Context, this *Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return ErrWaitKeepingInputs
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
				code:             ActivationCodeWaitingForInputsKeep,
				activationErrors: []error{ErrWaitKeepingInputs},
			},
		},
		{
			name: "component not activated, logger must be empty",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithOutputs("o1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						this.Logger().Println("This must not be logged, as component must not activate")
						return nil
					}),
				)
				require.NoError(t, err)
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodeNoInput),
			loggerAssertions: func(t *testing.T, output []byte) {
				assert.Empty(t, output)
			},
		},
		{
			name: "activated with error, with logging",
			getComponent: func() *Component {
				c, err := New("c1",
					WithInputs("i1"),
					WithActivationFunc(func(_ context.Context, this *Component) error {
						this.Logger().Println("This line must be logged")
						return errors.New("test error")
					}),
				)
				require.NoError(t, err)
				require.NoError(t, c.InputByName("i1").PutSignals(signal.New(123)))
				return c
			},
			wantActivationResult: NewActivationResult("c1", ActivationCodeReturnedError, errors.New("component returned an error: test error")),
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

			gotActivationResult := component.MaybeActivate(context.Background())
			assert.Equal(t, tt.wantActivationResult.Activated(), gotActivationResult.Activated())
			assert.Equal(t, tt.wantActivationResult.ComponentName(), gotActivationResult.ComponentName())
			assert.Equal(t, tt.wantActivationResult.Code(), gotActivationResult.Code())
			switch {
			case tt.wantActivationResult.IsError():
				require.EqualError(t, gotActivationResult.Err(), tt.wantActivationResult.Err().Error())
			case tt.wantActivationResult.IsPanic():
				// Panics used to land in the else branch below, so the expected
				// message was never compared with anything.
				require.EqualError(t, gotActivationResult.Err(), tt.wantActivationResult.Err().Error())
			default:
				assert.False(t, gotActivationResult.IsError())
				assert.False(t, gotActivationResult.IsPanic())
			}

			if tt.loggerAssertions != nil {
				tt.loggerAssertions(t, loggerOutput.Bytes())
			}
		})
	}
}

func TestComponent_MaybeActivate_HookFailures(t *testing.T) {
	t.Run("beforeActivation hook fails: ActivationCodeHookFailed, activated, error captured", func(t *testing.T) {
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(_ context.Context, this *Component) error { return nil }),
		)
		require.NoError(t, err)
		c.SetupHooks(func(h *Hooks) {
			h.BeforeActivation(func(_ context.Context, _ *Component) error {
				return errors.New("before hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		assert.True(t, result.Activated(), "activated so the drain clears its inputs")
		require.Error(t, result.Err())
		require.ErrorContains(t, result.Err(), "before hook error")
		assert.Len(t, result.Errors(), 1)
	})

	t.Run("afterActivation hook fails after an error: both errors accumulated", func(t *testing.T) {
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(_ context.Context, this *Component) error {
				return errors.New("component error")
			}),
		)
		require.NoError(t, err)
		c.SetupHooks(func(h *Hooks) {
			h.AfterActivation(func(_ context.Context, _ *ActivationContext) error {
				return errors.New("afterActivation hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		assert.Len(t, result.Errors(), 2)
		require.ErrorContains(t, result.Err(), "component error")
		assert.ErrorContains(t, result.Err(), "afterActivation hook error")
	})

	t.Run("afterActivation hook fails: ActivationCodeHookFailed, error accumulated", func(t *testing.T) {
		c, err := New("c1",
			WithInputs("i1"),
			WithActivationFunc(func(_ context.Context, this *Component) error { return nil }),
		)
		require.NoError(t, err)
		c.SetupHooks(func(h *Hooks) {
			h.AfterActivation(func(_ context.Context, _ *ActivationContext) error {
				return errors.New("afterActivation hook error")
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		assert.Equal(t, ActivationCodeHookFailed, result.Code())
		require.Error(t, result.Err())
		require.ErrorContains(t, result.Err(), "afterActivation hook error")
		assert.Len(t, result.Errors(), 1)
	})
}

func TestComponent_MaybeActivate_HookPanics(t *testing.T) {
	// Hooks run on the activation goroutine, where an escaped panic kills the
	// process, and each must fire once however the activation ended.
	ok := func(context.Context, *Component) error { return nil }

	tests := []struct {
		name       string
		activation ActivationFunc
		before     func(context.Context, *Component) error
		after      func(context.Context, *ActivationContext) error
		wantStage  string
		wantErrs   int
	}{
		{
			name:       "beforeActivation panics",
			activation: ok,
			before:     func(context.Context, *Component) error { panic("before boom") },
			wantStage:  "beforeActivation hook failed: panicked: before boom",
			wantErrs:   1,
		},
		{
			name:       "afterActivation panics after success",
			activation: ok,
			after:      func(context.Context, *ActivationContext) error { panic("after boom") },
			wantStage:  "afterActivation hook failed: panicked: after boom",
			wantErrs:   1,
		},
		{
			name:       "afterActivation fails after the function panicked",
			activation: func(context.Context, *Component) error { panic("function boom") },
			after:      func(context.Context, *ActivationContext) error { return errors.New("after failed") },
			wantStage:  "afterActivation hook failed: after failed",
			wantErrs:   2,
		},
		{
			name:       "afterActivation panics after the function panicked",
			activation: func(context.Context, *Component) error { panic("function boom") },
			after:      func(context.Context, *ActivationContext) error { panic("after boom") },
			wantStage:  "afterActivation hook failed: panicked: after boom",
			wantErrs:   2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New("c1", WithInputs("i1"), WithActivationFunc(tt.activation))
			require.NoError(t, err)

			afterCalls := 0
			c.SetupHooks(func(h *Hooks) {
				if tt.before != nil {
					h.BeforeActivation(tt.before)
				}
				h.AfterActivation(func(ctx context.Context, ac *ActivationContext) error {
					afterCalls++
					if tt.after != nil {
						return tt.after(ctx, ac)
					}
					return nil
				})
			})
			require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

			result := c.MaybeActivate(context.Background())

			assert.Equal(t, 1, afterCalls, "AfterActivation must run exactly once")
			assert.Equal(t, ActivationCodePanicked, result.Code())
			assert.True(t, result.IsPanic())
			assert.Len(t, result.Errors(), tt.wantErrs)
			require.ErrorContains(t, result.Err(), tt.wantStage)
			var panicErr *PanicError
			require.ErrorAs(t, result.Err(), &panicErr)
			assert.NotEmpty(t, panicErr.Stack)
		})
	}
}

func TestComponent_WithRetry(t *testing.T) {
	errBoom := errors.New("boom")
	// fails returns an activation function that fails the first n calls,
	// writing an output each time, and counts every call.
	fails := func(n int, calls *int, final error) ActivationFunc {
		return func(_ context.Context, this *Component) error {
			*calls++
			if err := this.OutputByName("o1").PutPayloads(*calls); err != nil {
				return err
			}
			if *calls <= n {
				return errBoom
			}
			return final
		}
	}

	tests := []struct {
		name      string
		attempts  int
		activate  func(calls *int) ActivationFunc
		ctx       func() context.Context
		wantCalls int
		wantCode  ActivationResultCode
		wantErrs  int
		wantOut   []any
	}{
		{
			name:      "succeeds on a later attempt, keeping only that attempt's outputs",
			attempts:  3,
			activate:  func(calls *int) ActivationFunc { return fails(2, calls, nil) },
			wantCalls: 3,
			wantCode:  ActivationCodeOK,
			wantOut:   []any{3},
		},
		{
			name:      "fails only when every attempt failed",
			attempts:  3,
			activate:  func(calls *int) ActivationFunc { return fails(3, calls, nil) },
			wantCalls: 3,
			wantCode:  ActivationCodeReturnedError,
			wantErrs:  3,
			wantOut:   []any{3},
		},
		{
			name:     "a panic on the first attempt is not retried",
			attempts: 3,
			activate: func(calls *int) ActivationFunc {
				return func(context.Context, *Component) error { *calls++; panic("first") }
			},
			wantCalls: 1,
			wantCode:  ActivationCodePanicked,
			wantErrs:  1,
		},
		{
			name:     "a panic after an error stops the retries",
			attempts: 3,
			activate: func(calls *int) ActivationFunc {
				return func(context.Context, *Component) error {
					*calls++
					if *calls == 2 {
						panic("second")
					}
					return errBoom
				}
			},
			wantCalls: 2,
			wantCode:  ActivationCodePanicked,
			wantErrs:  2,
		},
		{
			name:      "waiting for inputs is not retried",
			attempts:  3,
			activate:  func(calls *int) ActivationFunc { return fails(0, calls, ErrWaitKeepingInputs) },
			wantCalls: 1,
			wantCode:  ActivationCodeWaitingForInputsKeep,
			wantErrs:  1,
			wantOut:   []any{1},
		},
		{
			name:     "a canceled context stops the retries",
			attempts: 3,
			activate: func(calls *int) ActivationFunc { return fails(3, calls, nil) },
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			wantCalls: 1,
			wantCode:  ActivationCodeReturnedError,
			wantErrs:  1,
			wantOut:   []any{1},
		},
		{
			name:      "without retry an error fails at once",
			activate:  func(calls *int) ActivationFunc { return fails(1, calls, nil) },
			wantCalls: 1,
			wantCode:  ActivationCodeReturnedError,
			wantErrs:  1,
			wantOut:   []any{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls, before, after := 0, 0, 0
			opts := []Option{WithInputs("i1"), WithOutputs("o1"), WithActivationFunc(tt.activate(&calls))}
			if tt.attempts > 0 {
				opts = append(opts, WithRetry(tt.attempts))
			}
			c, err := New("c1", opts...)
			require.NoError(t, err)
			c.SetupHooks(func(h *Hooks) {
				h.BeforeActivation(func(context.Context, *Component) error { before++; return nil })
				h.AfterActivation(func(context.Context, *ActivationContext) error { after++; return nil })
			})
			require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}
			result := c.MaybeActivate(ctx)

			assert.Equal(t, tt.wantCalls, calls)
			assert.Equal(t, tt.wantCode, result.Code())
			assert.Len(t, result.Errors(), tt.wantErrs)
			assert.Equal(t, 1, before, "BeforeActivation fires once per activation")
			assert.Equal(t, 1, after, "AfterActivation fires once per activation")
			if tt.wantOut != nil {
				assert.Equal(t, tt.wantOut, c.OutputByName("o1").Signals().AllPayloads())
			}
		})
	}

	t.Run("each attempt's error is reported", func(t *testing.T) {
		calls := 0
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithRetry(2),
			WithActivationFunc(fails(2, &calls, nil)))
		require.NoError(t, err)
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		require.ErrorIs(t, result.Err(), errBoom)
		require.ErrorContains(t, result.Err(), "attempt 1 of 2: boom")
		assert.ErrorContains(t, result.Err(), "attempt 2 of 2: boom")
	})

	t.Run("outputs from before the activation survive a retry", func(t *testing.T) {
		calls := 0
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithRetry(2),
			WithActivationFunc(fails(1, &calls, nil)))
		require.NoError(t, err)
		require.NoError(t, c.OutputByName("o1").PutPayloads("earlier"))
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		assert.Equal(t, ActivationCodeOK, result.Code())
		assert.Equal(t, []any{"earlier", 2}, c.OutputByName("o1").Signals().AllPayloads())
	})

	t.Run("the output reset fires the port's hooks", func(t *testing.T) {
		// Documented behavior: the reset goes through Port.Clear and
		// PutSignalGroups, so port hooks see it.
		calls := 0
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithRetry(2),
			WithActivationFunc(fails(1, &calls, nil)))
		require.NoError(t, err)
		require.NoError(t, c.OutputByName("o1").PutPayloads("earlier"))

		clears, added := 0, 0
		c.OutputByName("o1").SetupHooks(func(h *port.Hooks) {
			h.OnClear(func(context.Context, *port.ClearContext) error { clears++; return nil })
			h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error { added++; return nil })
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		require.Equal(t, ActivationCodeOK, result.Code())
		assert.Equal(t, 1, clears, "one reset between the two attempts")
		assert.Equal(t, 3, added, "attempt 1 output, the earlier signal put back, attempt 2 output")
	})

	t.Run("an output empty before the activation is only cleared", func(t *testing.T) {
		calls := 0
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithRetry(2),
			WithActivationFunc(fails(1, &calls, nil)))
		require.NoError(t, err)

		var added []int
		c.OutputByName("o1").SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(_ context.Context, ac *port.SignalsAddedContext) error {
				added = append(added, len(ac.SignalsAdded))
				return nil
			})
		})
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		require.Equal(t, ActivationCodeOK, result.Code())
		assert.Equal(t, []int{1, 1}, added, "one output per attempt, no empty refill")
	})

	t.Run("a panic keeps the errors of earlier attempts", func(t *testing.T) {
		calls := 0
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithRetry(3),
			WithActivationFunc(func(context.Context, *Component) error {
				calls++
				if calls == 2 {
					panic("second")
				}
				return errBoom
			}))
		require.NoError(t, err)
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))

		result := c.MaybeActivate(context.Background())

		require.ErrorIs(t, result.Err(), errBoom)
		var panicErr *PanicError
		assert.ErrorAs(t, result.Err(), &panicErr)
	})

	t.Run("rejects fewer than one attempt", func(t *testing.T) {
		_, err := New("c1", WithRetry(0))
		require.ErrorContains(t, err, "retry attempts must be at least 1")
	})
}

func TestComponent_WithRetryIf(t *testing.T) {
	type call struct {
		attempt int
		err     error
	}
	// activate builds and runs a component whose activation function is f,
	// recording every call to the retry predicate, which answers with retry.
	activate := func(t *testing.T, ctx context.Context, attempts int, f ActivationFunc,
		retry func(ctx context.Context, attempt int, err error) bool,
	) (*ActivationResult, []call) {
		t.Helper()
		var calls []call
		c, err := New("c1", WithInputs("i1"), WithOutputs("o1"), WithActivationFunc(f), WithRetry(attempts),
			WithRetryIf(func(ctx context.Context, attempt int, err error) bool {
				calls = append(calls, call{attempt, err})
				return retry(ctx, attempt, err)
			}))
		require.NoError(t, err)
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New(1)))
		return c.MaybeActivate(ctx), calls
	}
	always := func(context.Context, int, error) bool { return true }
	// failing returns an activation function that returns "fail N" on call N.
	failing := func(runs *int) ActivationFunc {
		return func(context.Context, *Component) error {
			*runs++
			return fmt.Errorf("fail %d", *runs)
		}
	}

	t.Run("gets each attempt number and error, but not after the last attempt", func(t *testing.T) {
		runs := 0
		result, calls := activate(t, context.Background(), 3, failing(&runs), always)

		assert.Equal(t, 3, runs)
		assert.Equal(t, ActivationCodeReturnedError, result.Code())
		assert.Len(t, result.Errors(), 3)
		require.Len(t, calls, 2)
		assert.Equal(t, 1, calls[0].attempt)
		require.EqualError(t, calls[0].err, "fail 1")
		assert.Equal(t, 2, calls[1].attempt)
		require.EqualError(t, calls[1].err, "fail 2")
	})

	t.Run("returning false stops the retries", func(t *testing.T) {
		runs := 0
		result, calls := activate(t, context.Background(), 5, failing(&runs),
			func(_ context.Context, attempt int, _ error) bool { return attempt < 2 })

		assert.Equal(t, 2, runs)
		assert.Len(t, calls, 2)
		assert.Equal(t, ActivationCodeReturnedError, result.Code())
		assert.Len(t, result.Errors(), 2, "one error per attempt made")
	})

	t.Run("not called after a success", func(t *testing.T) {
		runs := 0
		f := func(ctx context.Context, this *Component) error {
			if runs == 1 {
				runs++
				return nil
			}
			return failing(&runs)(ctx, this)
		}
		result, calls := activate(t, context.Background(), 3, f, always)

		assert.Equal(t, ActivationCodeOK, result.Code())
		assert.Len(t, calls, 1, "called after the failed first attempt only")
	})

	t.Run("not called on a panic", func(t *testing.T) {
		result, calls := activate(t, context.Background(), 3,
			func(context.Context, *Component) error { panic("boom") }, always)

		assert.Equal(t, ActivationCodePanicked, result.Code())
		assert.Empty(t, calls)
	})

	t.Run("not called when waiting for inputs", func(t *testing.T) {
		result, calls := activate(t, context.Background(), 3,
			func(context.Context, *Component) error { return ErrWaitKeepingInputs }, always)

		assert.Equal(t, ActivationCodeWaitingForInputsKeep, result.Code())
		assert.Empty(t, calls)
	})

	t.Run("not called when the context is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		runs := 0
		result, calls := activate(t, ctx, 3, failing(&runs), always)

		assert.Equal(t, 1, runs)
		assert.Len(t, result.Errors(), 1)
		assert.Empty(t, calls)
	})

	// backoff waits a second before every retry, or gives up when ctx is done.
	backoff := func(ctx context.Context, _ int, _ error) bool {
		select {
		case <-time.After(time.Second):
			return true
		case <-ctx.Done():
			return false
		}
	}

	t.Run("backoff waits between attempts", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			runs := 0
			f := func(ctx context.Context, this *Component) error {
				if runs == 2 {
					runs++
					return nil
				}
				return failing(&runs)(ctx, this)
			}
			start := time.Now()
			result, _ := activate(t, context.Background(), 3, f, backoff)

			assert.Equal(t, ActivationCodeOK, result.Code())
			assert.Equal(t, 2*time.Second, time.Since(start))
		})
	})

	t.Run("backoff ends when the context is done", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
			defer cancel()
			runs := 0
			start := time.Now()
			result, calls := activate(t, ctx, 10, failing(&runs), backoff)

			assert.Equal(t, 2500*time.Millisecond, time.Since(start), "the third wait is cut short")
			assert.Equal(t, 3, runs)
			assert.Len(t, calls, 3)
			assert.Equal(t, ActivationCodeReturnedError, result.Code())
			assert.Len(t, result.Errors(), 3)
		})
	})

	t.Run("no attempt runs after the context ends during the wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		runs := 0
		result, calls := activate(t, ctx, 3, failing(&runs), func(context.Context, int, error) bool {
			cancel()
			return true
		})

		assert.Equal(t, 1, runs)
		assert.Len(t, calls, 1)
		assert.Equal(t, ActivationCodeReturnedError, result.Code())
	})

	t.Run("needs WithRetry with at least 2 attempts", func(t *testing.T) {
		_, err := New("c1", WithRetryIf(always))
		require.ErrorContains(t, err, "WithRetryIf needs WithRetry with at least 2 attempts")

		_, err = New("c1", WithRetryIf(always), WithRetry(1))
		require.ErrorContains(t, err, "WithRetryIf needs WithRetry with at least 2 attempts")

		_, err = New("c1", WithRetryIf(always), WithRetry(2))
		require.NoError(t, err, "option order does not matter")
	})

	t.Run("rejects a nil predicate", func(t *testing.T) {
		_, err := New("c1", WithRetry(2), WithRetryIf(nil))
		require.ErrorContains(t, err, "retry predicate must not be nil")
	})
}
