package fmesh

import (
	"errors"
	"testing"
	"time"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MultipleRun(t *testing.T) {
	t.Run("run result is initialized before each run", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		require.NoError(t, fm.AddComponents(
			mustNewComponent("bypass",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithDescription("Bypasses all signals"),
				component.WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				})),
		))

		for i := range 5 {
			require.NoError(t, fm.ComponentByName("bypass").InputByName("in").PutSignals(signal.New(i)))
			runResult, err := fm.Run()
			require.NoError(t, err)
			assert.NotNil(t, runResult)
			assert.Equal(t, 2, runResult.Cycles.Len())
			assert.Equal(t, 1, runResult.Cycles.Count(func(c *cycle.Cycle) bool {
				return c.HasActivatedComponents()
			}))
		}
	})

	t.Run("component state persists between runs", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		require.NoError(t, fm.AddComponents(
			mustNewComponent("counter",
				component.WithInputs("trigger"),
				component.WithOutputs("count"),
				component.WithDescription("Increments internal counter on each activation"),
				component.WithInitialState(func(state component.State) {
					state.Set("count", 0)
				}),
				component.WithActivationFunc(func(this *component.Component) error {
					count := this.State().Get("count").(int)
					count++
					this.State().Set("count", count)
					return this.OutputByName("count").PutSignals(signal.New(count))
				})),
		))

		require.NoError(t, fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go")))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult1.Cycles.Len())
		assert.Equal(t, 1, fm.ComponentByName("counter").State().Get("count"))

		require.NoError(t, fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go")))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult2.Cycles.Len())
		assert.Equal(t, 2, fm.ComponentByName("counter").State().Get("count"))

		require.NoError(t, fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go")))
		runResult3, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult3.Cycles.Len())
		assert.Equal(t, 3, fm.ComponentByName("counter").State().Get("count"))
	})

	t.Run("signals on output ports are cleared between runs", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		require.NoError(t, fm.AddComponents(
			mustNewComponent("producer",
				component.WithInputs("trigger"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("out").PutSignals(signal.New("data"))
				})),
			mustNewComponent("consumer",
				component.WithInputs("in"),
				component.WithOutputs("piped_out", "unpiped_out"),
				component.WithActivationFunc(func(this *component.Component) error {
					count := this.InputByName("in").Signals().Len()
					if err := this.OutputByName("piped_out").PutSignals(signal.New(count)); err != nil {
						return err
					}
					return this.OutputByName("unpiped_out").PutSignals(signal.New(count))
				})),
			mustNewComponent("final",
				component.WithInputs("in"),
				component.WithOutputs("result"),
				component.WithActivationFunc(func(this *component.Component) error {
					count := this.InputByName("in").Signals().Len()
					return this.OutputByName("result").PutSignals(signal.New(count))
				})),
		))

		require.NoError(t, fm.ComponentByName("producer").OutputByName("out").
			PipeTo(fm.ComponentByName("consumer").InputByName("in")))
		require.NoError(t, fm.ComponentByName("consumer").OutputByName("piped_out").
			PipeTo(fm.ComponentByName("final").InputByName("in")))

		// Run 1
		require.NoError(t, fm.ComponentByName("producer").InputByName("trigger").PutSignals(signal.New("go")))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		require.NotNil(t, runResult1)

		producerOutputSignals := fm.ComponentByName("producer").OutputByName("out").Signals().Len()
		t.Logf("Run 1: Producer output port has %d signals after run", producerOutputSignals)

		consumerUnpipedSignals1 := fm.ComponentByName("consumer").OutputByName("unpiped_out").Signals().Len()
		assert.Equal(t, 1, consumerUnpipedSignals1, "Run 1: Consumer unpiped output should have 1 signal")

		finalResultSignals1 := fm.ComponentByName("final").OutputByName("result").Signals().Len()
		assert.Equal(t, 1, finalResultSignals1, "Run 1: Final component should have 1 signal")

		// Run 2
		require.NoError(t, fm.ComponentByName("producer").InputByName("trigger").PutSignals(signal.New("go")))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		require.NotNil(t, runResult2)

		// Check that final component received exactly 1 signal in Run 2 (not accumulated from Run 1)
		count := fm.ComponentByName("final").OutputByName("result").Signals().FirstPayloadOrDefault(0).(int)
		assert.Equal(t, 1, count, "Final component should receive 1 signal per run, not accumulated signals from previous runs")

		producerOutputSignals2 := fm.ComponentByName("producer").OutputByName("out").Signals().Len()
		t.Logf("Run 2: Producer output port has %d signals after run", producerOutputSignals2)

		// Critical: Unpiped output should be cleared at start of Run 2, so it should only have 1 signal from this run
		consumerUnpipedSignals2 := fm.ComponentByName("consumer").OutputByName("unpiped_out").Signals().Len()
		assert.Equal(t, 1, consumerUnpipedSignals2, "Run 2: Consumer unpiped output should have 1 signal (not accumulated from Run 1)")

		finalResultSignals2 := fm.ComponentByName("final").OutputByName("result").Signals().Len()
		assert.Equal(t, 1, finalResultSignals2, "Run 2: Final component should have 1 signal (not accumulated from Run 1)")
	})

	t.Run("different cycle counts per run", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		require.NoError(t, fm.AddComponents(
			mustNewComponent("repeater",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					count := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
					if count > 0 {
						return this.OutputByName("out").PutSignals(signal.New(count - 1))
					}
					return nil
				})),
		))

		require.NoError(t, fm.ComponentByName("repeater").OutputByName("out").
			PipeTo(fm.ComponentByName("repeater").InputByName("in")))

		require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(2)))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 4, runResult1.Cycles.Len(), "Run 1: expected 4 cycles (3 with activation + 1 empty)")

		require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(4)))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 6, runResult2.Cycles.Len(), "Run 2: expected 6 cycles (5 with activation + 1 empty)")

		require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(0)))
		runResult3, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult3.Cycles.Len(), "Run 3: expected 2 cycles (1 with activation + 1 empty)")
	})

	t.Run("hooks execute correctly across multiple runs", func(t *testing.T) {
		beforeRunCount := 0
		afterRunCount := 0
		beforeCycleCount := 0
		afterCycleCount := 0

		fm := mustNewFMesh("test fm")
		fm.SetupHooks(func(h *Hooks) {
			h.BeforeRun(func(fm *FMesh) error {
				beforeRunCount++
				return nil
			})
			h.AfterRun(func(fm *FMesh) error {
				afterRunCount++
				return nil
			})
			h.BeforeCycle(func(ctx *CycleContext) error {
				beforeCycleCount++
				return nil
			})
			h.AfterCycle(func(ctx *CycleContext) error {
				afterCycleCount++
				return nil
			})
		})
		require.NoError(t, fm.AddComponents(
			mustNewComponent("simple",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				})),
		))

		require.NoError(t, fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()
		require.NoError(t, err)

		require.NoError(t, fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(2)))
		_, err = fm.Run()
		require.NoError(t, err)

		require.NoError(t, fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(3)))
		_, err = fm.Run()
		require.NoError(t, err)

		assert.Equal(t, 3, beforeRunCount, "BeforeRun should be called 3 times")
		assert.Equal(t, 3, afterRunCount, "AfterRun should be called 3 times")
		assert.Equal(t, 6, beforeCycleCount, "BeforeCycle should be called 6 times (2 per run)")
		assert.Equal(t, 6, afterCycleCount, "AfterCycle should be called 6 times (2 per run)")
	})

	t.Run("mesh with error handling strategy stops on error", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
			CyclesLimit:           0,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("faulty",
				component.WithInputs("trigger"),
				component.WithActivationFunc(func(this *component.Component) error {
					return errors.New("intentional error")
				})),
		))

		require.NoError(t, fm.ComponentByName("faulty").InputByName("trigger").PutSignals(signal.New("go")))
		runResult1, err := fm.Run()
		require.Error(t, err, "Run 1 should fail")
		assert.True(t, runResult1.Cycles.Last().HasActivationErrors())
	})

	t.Run("runtime info duration is per run", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		require.NoError(t, fm.AddComponents(
			mustNewComponent("sleeper",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					sleepDuration := this.InputByName("in").Signals().FirstPayloadOrDefault(time.Duration(0)).(time.Duration)
					time.Sleep(sleepDuration)
					return nil
				})),
		))

		require.NoError(t, fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(10*time.Millisecond)))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		duration1 := runResult1.Duration()
		assert.Positive(t, duration1)
		t.Logf("Run 1 duration (10ms sleep): %d nanoseconds", duration1)

		require.NoError(t, fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(50*time.Millisecond)))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		duration2 := runResult2.Duration()
		assert.Positive(t, duration2)
		t.Logf("Run 2 duration (50ms sleep): %d nanoseconds", duration2)

		assert.Greater(t, duration2, duration1)
		assert.GreaterOrEqual(t, duration1, int64(10*time.Millisecond))
		assert.GreaterOrEqual(t, duration2, int64(50*time.Millisecond))
		assert.Less(t, duration2, int64(100*time.Millisecond), "Run 2 duration should not be accumulated from Run 1")
	})

	t.Run("cycles history limit trims retained cycles to a sliding window", func(t *testing.T) {
		newRepeaterMesh := func(opts ...Option) *FMesh {
			fm := mustNewFMesh("test fm", opts...)
			require.NoError(t, fm.AddComponents(
				mustNewComponent("repeater",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(func(this *component.Component) error {
						count := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
						if count > 0 {
							return this.OutputByName("out").PutSignals(signal.New(count - 1))
						}
						return nil
					})),
			))
			require.NoError(t, fm.ComponentByName("repeater").OutputByName("out").
				PipeTo(fm.ComponentByName("repeater").InputByName("in")))
			return fm
		}

		t.Run("limit smaller than run length", func(t *testing.T) {
			fm := newRepeaterMesh(WithCyclesHistoryLimit(3))

			// Observe retention during the run, not only in the returned result:
			// the retained history must never exceed the limit between cycles.
			var maxObservedLen int
			var observedCycles int
			fm.SetupHooks(func(h *Hooks) {
				h.AfterCycle(func(ctx *CycleContext) error {
					observedCycles++
					if l := ctx.FMesh.runtimeInfo.Cycles.Len(); l > maxObservedLen {
						maxObservedLen = l
					}
					return nil
				})
			})

			require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(5)))
			runResult, err := fm.Run()
			require.NoError(t, err)

			// 5 -> 4 -> 3 -> 2 -> 1 -> 0 (6 activated cycles) + 1 empty cycle = 7 total cycles (M = 7)
			const totalCycles = 7
			assert.Equal(t, 3, runResult.Cycles.Len())
			assert.Equal(t, totalCycles, runResult.Cycles.Last().Number())
			assert.Equal(t, totalCycles-3+1, runResult.Cycles.First().Number())

			// The mesh executed more cycles than it retained, and retention never exceeded the limit mid-run.
			assert.Equal(t, totalCycles, observedCycles)
			assert.LessOrEqual(t, maxObservedLen, 3)
		})

		t.Run("limit larger than run length retains full history", func(t *testing.T) {
			fm := newRepeaterMesh(WithCyclesHistoryLimit(100))
			require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(2)))
			runResult, err := fm.Run()
			require.NoError(t, err)

			assert.Equal(t, 4, runResult.Cycles.Len())
			assert.Equal(t, 1, runResult.Cycles.First().Number())
			assert.Equal(t, 4, runResult.Cycles.Last().Number())
		})
	})

	t.Run("NoInput activation results are never recorded", func(t *testing.T) {
		newMeshWithIdleComponent := func(opts ...Option) *FMesh {
			fm := mustNewFMesh("test fm", opts...)
			require.NoError(t, fm.AddComponents(
				mustNewComponent("repeater",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(func(this *component.Component) error {
						count := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
						if count > 0 {
							return this.OutputByName("out").PutSignals(signal.New(count - 1))
						}
						return nil
					})),
				// idle never receives any input, so it always reports ActivationCodeNoInput
				mustNewComponent("idle",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(func(this *component.Component) error {
						return nil
					})),
			))
			require.NoError(t, fm.ComponentByName("repeater").OutputByName("out").
				PipeTo(fm.ComponentByName("repeater").InputByName("in")))
			return fm
		}

		t.Run("idle component's results never appear, mesh still drains and stops correctly", func(t *testing.T) {
			fm := newMeshWithIdleComponent()
			require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(2)))
			runResult, err := fm.Run()
			require.NoError(t, err)

			// 2 -> 1 -> 0 (3 activated cycles) + 1 empty cycle = 4 total cycles
			assert.Equal(t, 4, runResult.Cycles.Len())

			for _, c := range runResult.Cycles.All() {
				assert.Nil(t, c.ActivationResults().ByName("idle"), "cycle #%d", c.Number())

				repeaterResult := c.ActivationResults().ByName("repeater")
				if repeaterResult == nil {
					// repeater itself had no input in the final cycle: its NoInput result is not recorded either.
					continue
				}
				assert.True(t, repeaterResult.Activated())
				assert.NotEqual(t, component.ActivationCodeNoInput, repeaterResult.Code())
			}
		})

		t.Run("combined with a cycles history limit", func(t *testing.T) {
			fm := newMeshWithIdleComponent(WithCyclesHistoryLimit(2))
			require.NoError(t, fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(5)))
			runResult, err := fm.Run()
			require.NoError(t, err)

			assert.Equal(t, 2, runResult.Cycles.Len())
			for _, c := range runResult.Cycles.All() {
				assert.Nil(t, c.ActivationResults().ByName("idle"), "cycle #%d", c.Number())
			}
		})
	})
}
