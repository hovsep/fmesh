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
		fm := New("test fm").
			AddComponents(
				component.New("bypass").
					WithDescription("Bypasses all signals").
					AddInputs("in").
					AddOutputs("out").WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				}))

		for i := 0; i < 5; i++ {
			fm.ComponentByName("bypass").InputByName("in").PutSignals(signal.New(i))
			runResult, err := fm.Run()
			require.NoError(t, err)
			assert.NotNil(t, runResult)
			assert.Equal(t, 2, runResult.Cycles.Len())
			assert.Equal(t, 1, runResult.Cycles.CountMatch(func(c *cycle.Cycle) bool {
				return c.HasActivatedComponents()
			}))
		}
	})

	t.Run("component state persists between runs", func(t *testing.T) {
		fm := New("test fm").
			AddComponents(
				component.New("counter").
					WithDescription("Increments internal counter on each activation").
					AddInputs("trigger").
					AddOutputs("count").
					WithInitialState(func(state component.State) {
						state.Set("count", 0)
					}).
					WithActivationFunc(func(this *component.Component) error {
						count := this.State().Get("count").(int)
						count++
						this.State().Set("count", count)
						this.OutputByName("count").PutSignals(signal.New(count))
						return nil
					}))

		fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go"))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult1.Cycles.Len())
		assert.Equal(t, 1, fm.ComponentByName("counter").State().Get("count"))

		fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go"))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult2.Cycles.Len())
		assert.Equal(t, 2, fm.ComponentByName("counter").State().Get("count"))

		fm.ComponentByName("counter").InputByName("trigger").PutSignals(signal.New("go"))
		runResult3, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult3.Cycles.Len())
		assert.Equal(t, 3, fm.ComponentByName("counter").State().Get("count"))
	})

	t.Run("signals on output ports are cleared between runs", func(t *testing.T) {
		fm := New("test fm").
			AddComponents(
				component.New("producer").
					AddInputs("trigger").
					AddOutputs("out").
					WithActivationFunc(func(this *component.Component) error {
						this.OutputByName("out").PutSignals(signal.New("data"))
						return nil
					}),
				component.New("consumer").
					AddInputs("in").
					AddOutputs("out").
					WithActivationFunc(func(this *component.Component) error {
						count := this.InputByName("in").Signals().Len()
						this.OutputByName("out").PutSignals(signal.New(count))
						return nil
					}))

		fm.ComponentByName("producer").OutputByName("out").
			PipeTo(fm.ComponentByName("consumer").InputByName("in"))

		fm.ComponentByName("producer").InputByName("trigger").PutSignals(signal.New("go"))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		require.NoError(t, runResult1.Cycles.ChainableErr())

		producerOutputSignals := fm.ComponentByName("producer").OutputByName("out").Signals().Len()
		t.Logf("Run 1: Producer output port has %d signals after run", producerOutputSignals)

		fm.ComponentByName("producer").InputByName("trigger").PutSignals(signal.New("go"))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		require.NoError(t, runResult2.Cycles.ChainableErr())

		count := fm.ComponentByName("consumer").OutputByName("out").Signals().FirstPayloadOrDefault(0).(int)
		assert.Equal(t, 1, count, "Consumer should receive 1 signal per run, not accumulated signals from previous runs")

		producerOutputSignals2 := fm.ComponentByName("producer").OutputByName("out").Signals().Len()
		t.Logf("Run 2: Producer output port has %d signals after run", producerOutputSignals2)
	})

	t.Run("mesh cannot run again after error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
			CyclesLimit:           0,
		}).AddComponents(
			component.New("faulty").
				AddInputs("trigger").
				WithActivationFunc(func(this *component.Component) error {
					return errors.New("intentional error")
				}))

		fm.ComponentByName("faulty").InputByName("trigger").PutSignals(signal.New("go"))
		runResult1, err := fm.Run()
		require.Error(t, err, "Run 1 should fail")
		assert.True(t, runResult1.Cycles.Last().HasActivationErrors())
		assert.True(t, fm.HasChainableErr(), "Mesh should have chainable error after failed run")

		fm.ComponentByName("faulty").InputByName("trigger").PutSignals(signal.New("go"))
		_, err = fm.Run()
		require.Error(t, err, "Run 2 should fail immediately - mesh has chainable error")
		assert.True(t, fm.HasChainableErr(), "Chainable error should persist")
		assert.ErrorContains(t, err, "error in fmesh")
	})

	t.Run("different cycle counts per run", func(t *testing.T) {
		fm := New("test fm").
			AddComponents(
				component.New("repeater").
					AddInputs("in").
					AddOutputs("out").
					WithActivationFunc(func(this *component.Component) error {
						count := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
						if count > 0 {
							this.OutputByName("out").PutSignals(signal.New(count - 1))
						}
						return nil
					}))

		fm.ComponentByName("repeater").OutputByName("out").
			PipeTo(fm.ComponentByName("repeater").InputByName("in"))

		fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(2))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 4, runResult1.Cycles.Len(), "Run 1: expected 4 cycles (3 with activation + 1 empty)")

		fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(4))
		runResult2, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 6, runResult2.Cycles.Len(), "Run 2: expected 6 cycles (5 with activation + 1 empty)")

		fm.ComponentByName("repeater").InputByName("in").PutSignals(signal.New(0))
		runResult3, err := fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, runResult3.Cycles.Len(), "Run 3: expected 2 cycles (1 with activation + 1 empty)")
	})

	t.Run("hooks execute correctly across multiple runs", func(t *testing.T) {
		beforeRunCount := 0
		afterRunCount := 0
		cycleBeginCount := 0
		cycleEndCount := 0

		fm := New("test fm").
			SetupHooks(func(h *Hooks) {
				h.BeforeRun(func(fm *FMesh) error {
					beforeRunCount++
					return nil
				})
				h.AfterRun(func(fm *FMesh) error {
					afterRunCount++
					return nil
				})
				h.CycleBegin(func(ctx *CycleContext) error {
					cycleBeginCount++
					return nil
				})
				h.CycleEnd(func(ctx *CycleContext) error {
					cycleEndCount++
					return nil
				})
			}).
			AddComponents(
				component.New("simple").
					AddInputs("in").
					AddOutputs("out").
					WithActivationFunc(func(this *component.Component) error {
						return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
					}))

		fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()
		require.NoError(t, err)

		fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(2))
		_, err = fm.Run()
		require.NoError(t, err)

		fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(3))
		_, err = fm.Run()
		require.NoError(t, err)

		assert.Equal(t, 3, beforeRunCount, "BeforeRun should be called 3 times")
		assert.Equal(t, 3, afterRunCount, "AfterRun should be called 3 times")
		assert.Equal(t, 6, cycleBeginCount, "CycleBegin should be called 6 times (2 per run)")
		assert.Equal(t, 6, cycleEndCount, "CycleEnd should be called 6 times (2 per run)")
	})

	t.Run("mesh cannot run again after panic", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			ErrorHandlingStrategy: StopOnFirstPanic,
			CyclesLimit:           0,
		}).AddComponents(
			component.New("panicky").
				AddInputs("trigger").
				WithActivationFunc(func(this *component.Component) error {
					panic("intentional panic")
				}))

		fm.ComponentByName("panicky").InputByName("trigger").PutSignals(signal.New("go"))
		runResult1, err := fm.Run()
		require.Error(t, err, "Run 1 should fail with panic")
		assert.True(t, runResult1.Cycles.Last().HasActivationPanics())
		assert.True(t, fm.HasChainableErr(), "Mesh should have chainable error after panic")

		fm.ComponentByName("panicky").InputByName("trigger").PutSignals(signal.New("go"))
		_, err = fm.Run()
		require.Error(t, err, "Run 2 should fail immediately - mesh has chainable error")
		assert.True(t, fm.HasChainableErr(), "Chainable error should persist")
		assert.ErrorContains(t, err, "error in fmesh")
	})

	t.Run("runtime info duration is per run", func(t *testing.T) {
		fm := New("test fm").
			AddComponents(
				component.New("sleeper").
					AddInputs("in").
					AddOutputs("out").
					WithActivationFunc(func(this *component.Component) error {
						sleepDuration := this.InputByName("in").Signals().FirstPayloadOrDefault(time.Duration(0)).(time.Duration)
						time.Sleep(sleepDuration)
						return nil
					}))

		fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(10 * time.Millisecond))
		runResult1, err := fm.Run()
		require.NoError(t, err)
		duration1 := runResult1.Duration()
		assert.Positive(t, duration1)
		t.Logf("Run 1 duration (10ms sleep): %d nanoseconds", duration1)

		fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(50 * time.Millisecond))
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
}
