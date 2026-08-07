package profiler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiler(t *testing.T) {
	p := New()

	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	require.NoError(t, fm.AddComponents(
		testutil.MustComponent("producer",
			component.WithInputs("i1"),
			component.WithOutputs("o1"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				return this.OutputByName("o1").PutPayloads(1)
			})),
		testutil.MustComponent("consumer",
			component.WithInputs("i1"),
			component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
	))

	require.NoError(t, fm.Components().ByName("producer").OutputByName("o1").
		PipeTo(fm.Components().ByName("consumer").InputByName("i1")))
	require.NoError(t, fm.Components().ByName("producer").InputByName("i1").
		PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
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
	p := New()

	assert.Zero(t, p.Runs().Count)
	assert.Zero(t, p.Cycles().Count)
	assert.Empty(t, p.Components())
	assert.Empty(t, p.TopN(3))
	assert.Contains(t, p.Report(), "runs:   0", "the header renders without any data behind it")
}

func TestProfiler_RanksByActivationCount(t *testing.T) {
	// TestProfiler times two components that each activate exactly once, so it
	// never sees TopN actually rank anything.
	p := New()
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	looper := testutil.MustComponent("looper",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			left, _ := this.State().GetOrDefault("left", 3).(int)
			if left <= 1 {
				return nil
			}
			this.State().Set("left", left-1)
			return this.OutputByName("o1").PutPayloads(left - 1)
		}))
	require.NoError(t, looper.LoopbackPipe("o1", "i1"))

	oneShot := testutil.MustComponent("one-shot",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(looper, oneShot))
	require.NoError(t, looper.InputByName("i1").PutSignals(signal.New("go")))
	require.NoError(t, oneShot.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
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
	p := New()
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
	p := New()

	fm, err := fmesh.New("m",
		fmesh.WithPlugins(p),
		fmesh.WithErrorHandlingStrategy(fmesh.IgnoreAll))
	require.NoError(t, err)

	c := testutil.MustComponent("doomed",
		component.WithInputs("i1"),
		component.WithHooks(func(hooks *component.Hooks) {
			hooks.BeforeActivation(func(context.Context, *component.Component) error {
				return errors.New("refused")
			})
		}),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	assert.Empty(t, p.Components(), "an activation that never began is not timed")
}
