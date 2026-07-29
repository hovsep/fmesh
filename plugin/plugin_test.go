package plugin

import (
	"errors"
	"testing"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiler(t *testing.T) {
	p := NewProfiler()

	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	require.NoError(t, fm.AddComponents(
		mustComponent("producer",
			component.WithInputs("i1"),
			component.WithOutputs("o1"),
			component.WithActivationFunc(func(this *component.Component) error {
				return this.OutputByName("o1").PutPayloads(1)
			})),
		mustComponent("consumer",
			component.WithInputs("i1"),
			component.WithActivationFunc(func(*component.Component) error { return nil })),
	))

	require.NoError(t, fm.Components().ByName("producer").OutputByName("o1").
		PipeTo(fm.Components().ByName("consumer").InputByName("i1")))
	require.NoError(t, fm.Components().ByName("producer").InputByName("i1").
		PutSignals(signal.New("go")))

	_, err = fm.Run()
	require.NoError(t, err)

	assert.Equal(t, 1, p.Runs().Count)
	assert.Positive(t, p.Cycles().Count, "the run took at least one cycle")

	stats := p.Components()
	require.Len(t, stats, 2, "both components timed, without either knowing")

	byName := map[string]ComponentStat{}
	for _, s := range stats {
		byName[s.Component] = s
	}
	assert.Equal(t, 1, byName["producer"].Count)
	assert.Equal(t, 1, byName["consumer"].Count)
	assert.GreaterOrEqual(t, byName["producer"].Max, byName["producer"].Min)

	assert.Contains(t, p.Report(), "producer")

	t.Run("TopN ranks by how often, not how slow", func(t *testing.T) {
		top := p.TopN(1)
		require.Len(t, top, 1)
		assert.Equal(t, 1, top[0].Count)
	})

	t.Run("TopN clamps to what was measured", func(t *testing.T) {
		assert.Len(t, p.TopN(100), 2, "asking for more than exists returns everything")
		assert.Empty(t, p.TopN(0))
		assert.Empty(t, p.TopN(-1), "a negative n returns nothing rather than panicking")
	})

	t.Run("an empty stat has no average", func(t *testing.T) {
		assert.Zero(t, Stat{}.Avg())
	})

	t.Run("Reset discards everything measured", func(t *testing.T) {
		p.Reset()

		assert.Zero(t, p.Runs().Count)
		assert.Zero(t, p.Cycles().Count)
		assert.Empty(t, p.Components())
	})
}

func TestProfiler_NothingMeasured(t *testing.T) {
	p := NewProfiler()

	assert.Zero(t, p.Runs().Count)
	assert.Zero(t, p.Cycles().Count)
	assert.Empty(t, p.Components())
	assert.Empty(t, p.TopN(3))
	assert.Contains(t, p.Report(), "runs:   0", "the header renders without any data behind it")
}

func TestProfiler_RanksByActivationCount(t *testing.T) {
	// TestProfiler times two components that each activate exactly once, so it
	// never sees TopN actually rank anything.
	p := NewProfiler()
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	looper := mustComponent("looper",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(this *component.Component) error {
			left, _ := this.State().GetOrDefault("left", 3).(int)
			if left <= 1 {
				return nil
			}
			this.State().Set("left", left-1)
			return this.OutputByName("o1").PutPayloads(left - 1)
		}))
	require.NoError(t, looper.LoopbackPipe("o1", "i1"))

	oneShot := mustComponent("one-shot",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(*component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(looper, oneShot))
	require.NoError(t, looper.InputByName("i1").PutSignals(signal.New("go")))
	require.NoError(t, oneShot.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run()
	require.NoError(t, err)

	top := p.TopN(2)
	require.Len(t, top, 2)
	assert.Equal(t, "looper", top[0].Component, "the busiest component comes first")
	assert.Greater(t, top[0].Count, top[1].Count)
}

func TestProfiler_TiesBreakOnName(t *testing.T) {
	// Both sorts fall back to the component name so a report cannot reorder
	// itself between runs. Real durations never tie, so the stats are planted
	// directly -- there is no other way to reach the tiebreak deterministically.
	p := NewProfiler()
	p.components = map[string]Stat{
		"charlie": {Count: 2, Total: time.Second},
		"alpha":   {Count: 2, Total: time.Second},
		"bravo":   {Count: 9, Total: time.Millisecond},
	}

	byTotal := make([]string, 0, 3)
	for _, s := range p.Components() {
		byTotal = append(byTotal, s.Component)
	}
	assert.Equal(t, []string{"alpha", "charlie", "bravo"}, byTotal,
		"slowest total first, equal totals in name order")

	byCount := make([]string, 0, 3)
	for _, s := range p.TopN(3) {
		byCount = append(byCount, s.Component)
	}
	assert.Equal(t, []string{"bravo", "alpha", "charlie"}, byCount,
		"busiest first, equal counts in name order")
}

func TestProfiler_ActivationThatNeverBegan(t *testing.T) {
	// AfterActivation is a finally block: a component whose BeforeActivation hook
	// fails still reaches it, with no start time recorded. Registering the failing
	// hook before the plugin arrives puts it ahead of the profiler's own hook in
	// the group, so the profiler never sees the start.
	p := NewProfiler()

	fm, err := fmesh.New("m",
		fmesh.WithPlugins(p),
		fmesh.WithErrorHandlingStrategy(fmesh.IgnoreAll))
	require.NoError(t, err)

	c := mustComponent("doomed",
		component.WithInputs("i1"),
		component.WithHooks(func(hooks *component.Hooks) {
			hooks.BeforeActivation(func(*component.Component) error {
				return errors.New("refused")
			})
		}),
		component.WithActivationFunc(func(*component.Component) error { return nil }))
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run()
	require.NoError(t, err)

	assert.Empty(t, p.Components(), "an activation that never began is not timed")
}

func TestAutowire(t *testing.T) {
	t.Run("wires by convention regardless of the order components arrive", func(t *testing.T) {
		// The consumer is added first, so it can only be wired by the second
		// component's arrival -- which is the half that a one-directional
		// implementation silently gets wrong.
		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowirePrefixed("env_")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			mustComponent("body",
				component.WithInputs("env_sun_uvi"),
				component.WithActivationFunc(recordArrival("env_sun_uvi"))),
		))
		require.NoError(t, fm.AddComponents(
			mustComponent("sun",
				component.WithInputs("i1"),
				component.WithOutputs("uvi"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("uvi").PutPayloads(7)
				})),
		))

		require.NoError(t, fm.Components().ByName("sun").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run()
		require.NoError(t, err)

		assert.Equal(t, 7, fm.Components().ByName("body").State().Get("arrived"),
			"the sun reached the body with no wiring written")
	})

	t.Run("AutowireBroadcast feeds every same-named input", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowireBroadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			mustComponent("clock",
				component.WithInputs("i1"),
				component.WithOutputs("time"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("time").PutPayloads(1)
				})),
			mustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
			mustComponent("lung",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
		))

		require.NoError(t, fm.Components().ByName("clock").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run()
		require.NoError(t, err)

		assert.Equal(t, 1, fm.Components().ByName("heart").State().Get("arrived"))
		assert.Equal(t, 1, fm.Components().ByName("lung").State().Get("arrived"))
	})

	t.Run("declining to name a port wires nothing", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowireBroadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			mustComponent("clock",
				component.WithOutputs("not_time"),
				component.WithActivationFunc(func(*component.Component) error { return nil })),
			mustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
		))

		assert.Empty(t, fm.Components().ByName("clock").OutputByName("not_time").Pipes().Len(),
			"a declined port is left unwired")
	})

	t.Run("a plugin with no naming rule fails construction", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(&Autowire{}))

		require.Error(t, err)
		require.ErrorContains(t, err, "Name must be set")
		assert.Nil(t, fm)
	})

	t.Run("naming an input the destination does not have wires nothing", func(t *testing.T) {
		// Distinct from declining: the rule does name a port, there is just no
		// such input on this component.
		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowireBroadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			mustComponent("clock",
				component.WithOutputs("time"),
				component.WithActivationFunc(func(*component.Component) error { return nil })),
			mustComponent("deaf",
				component.WithInputs("not_time"),
				component.WithActivationFunc(func(*component.Component) error { return nil })),
		))

		assert.Zero(t, fm.Components().ByName("clock").OutputByName("time").Pipes().Len())
	})

	t.Run("two conventions agreeing on a pair wire it once", func(t *testing.T) {
		// Pipes are not deduplicated and a port flushes once per pipe, so a
		// second identical pipe would deliver every signal twice.
		everythingIsTime := &Autowire{
			PluginName: "autowire:everything-is-time",
			Name:       func(*component.Component, *port.Port) string { return "time" },
		}

		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowireBroadcast("time"), everythingIsTime))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			mustComponent("clock",
				component.WithInputs("i1"),
				component.WithOutputs("time"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("time").PutPayloads(1)
				})),
			mustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(func(this *component.Component) error {
					this.State().Set("received", this.InputByName("time").Signals().Len())
					return nil
				})),
		))

		assert.Equal(t, 1, fm.Components().ByName("clock").OutputByName("time").Pipes().Len())

		require.NoError(t, fm.Components().ByName("clock").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run()
		require.NoError(t, err)

		assert.Equal(t, 1, fm.Components().ByName("heart").State().Get("received"),
			"the tick arrived once, not once per convention")
	})

	t.Run("a failing pipe fails AddComponents", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(AutowireBroadcast("time")))
		require.NoError(t, err)

		heart := mustComponent("heart",
			component.WithInputs("time"),
			component.WithActivationFunc(func(*component.Component) error { return nil }))
		heart.InputByName("time").SetupHooks(func(hooks *port.Hooks) {
			hooks.OnInboundPipe(func(*port.InboundPipeContext) error {
				return errors.New("refused")
			})
		})

		err = fm.AddComponents(
			mustComponent("clock",
				component.WithOutputs("time"),
				component.WithActivationFunc(func(*component.Component) error { return nil })),
			heart,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "refused")
	})
}

// recordArrival notes what turned up on a port, since the mesh drains ports once
// a cycle is done and nothing is left to inspect after Run returns.
func recordArrival(portName string) component.ActivationFunc {
	return func(this *component.Component) error {
		this.State().Set("arrived", this.InputByName(portName).Signals().FirstPayloadOrNil())
		return nil
	}
}

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}
