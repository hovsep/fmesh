package profiler

import (
	"context"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countdown emits n-1 on its output until it reaches one, so a mesh runs for a
// predictable number of cycles.
func countdown(name string, from int) *component.Component {
	c := testutil.MustComponent(name,
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			left, _ := this.State().GetOrDefault("left", from).(int)
			if left <= 1 {
				return nil
			}
			this.State().Set("left", left-1)
			return this.OutputByName("o1").PutPayloads(left - 1)
		}))
	if err := c.LoopbackPipe("o1", "i1"); err != nil {
		panic(err)
	}
	return c
}

func TestProfiler_Timeline(t *testing.T) {
	p := New(ModeTimeline)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	c := countdown("looper", 3)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	records := p.Timeline()
	require.NotEmpty(t, records)

	for i, r := range records {
		assert.Equal(t, 1, r.Run, "one run, so every record belongs to it")
		assert.Equal(t, i+1, r.Number, "cycle numbers are 1-based and contiguous")
		assert.Positive(t, r.Duration, "a cycle that reached AfterCycle has a duration")
		assert.Zero(t, r.Errors)
		assert.Zero(t, r.Panics)
	}
	assert.Equal(t, 1, records[0].Activations)
	assert.Contains(t, p.Report(), "timeline:")

	t.Run("Reset discards the timeline", func(t *testing.T) {
		p.Reset()
		assert.Empty(t, p.Timeline())
	})
}

func TestProfiler_TimelineAttributesDrainToItsOwnCycle(t *testing.T) {
	// The load-bearing case. A cycle's drain runs after its AfterCycle hook, so
	// the naive implementation charges every delivery to the next cycle -- which
	// silently shifts the whole chart one step to the right.
	p := New(ModeTimeline, ModeThroughput)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	producer := testutil.MustComponent("producer",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("o1").PutPayloads(1, 2)
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("o1").PipeTo(consumer.InputByName("i1")))
	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	records := p.Timeline()
	// Producer, then consumer, then the idle cycle that ends the run.
	require.Len(t, records, 3)
	assert.Equal(t, 2, records[0].SignalsMoved,
		"the drain that followed cycle 1 belongs to cycle 1, not cycle 2")
	assert.Zero(t, records[1].SignalsMoved, "the consumer sends nothing on")
	assert.Zero(t, records[2].SignalsMoved,
		"the last cycle of a run is never drained, so it always reports nothing moved")
}

func TestProfiler_TimelineIgnoresSignalsOutsideACycle(t *testing.T) {
	// Seeding before a run and putting after one both fire port hooks with no
	// cycle open. Charging either to a record would inflate the first or last
	// cycle of the chart.
	p := New(ModeTimeline, ModeThroughput)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	producer := testutil.MustComponent("producer",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("o1").PipeTo(consumer.InputByName("i1")))

	// Flushed by hand before the run: a real delivery, but no cycle owns it.
	require.NoError(t, producer.OutputByName("o1").PutPayloads(1, 2, 3))
	require.NoError(t, producer.OutputByName("o1").Flush(context.Background()))

	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))
	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	// And again after it.
	require.NoError(t, producer.OutputByName("o1").PutPayloads(9))
	require.NoError(t, producer.OutputByName("o1").Flush(context.Background()))

	for _, r := range p.Timeline() {
		assert.Zero(t, r.SignalsMoved, "no cycle moved anything in this mesh")
	}
	assert.Equal(t, 4, p.Pipes()[0].Signals, "the pipe still counted both hand flushes")
}

func TestProfiler_TimelineRunIndex(t *testing.T) {
	// Cycle numbers restart at 1 on every run while the timeline accumulates, so
	// without Run the x-axis repeats itself.
	p := New(ModeTimeline)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	c := countdown("looper", 2)
	require.NoError(t, fm.AddComponents(c))

	for range 2 {
		c.State().Delete("left")
		require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))
		_, err = fm.Run(context.Background())
		require.NoError(t, err)
	}

	records := p.Timeline()
	require.NotEmpty(t, records)
	assert.Equal(t, 1, records[0].Run)
	assert.Equal(t, 1, records[0].Number)
	assert.Equal(t, 2, records[len(records)-1].Run, "the second run is distinguishable")

	// Cycle 1 shows up twice, once per run, and only Run tells them apart.
	firstCycles := 0
	for _, r := range records {
		if r.Number == 1 {
			firstCycles++
		}
	}
	assert.Equal(t, 2, firstCycles)
}

func TestProfiler_TimelineLimit(t *testing.T) {
	// Unbounded is a leak for a long-running mesh, so the timeline evicts. It
	// drops the oldest in chunks, which keeps the cap a ceiling rather than an
	// exact retained count.
	// Timing is on only so the test can compare against the true cycle count
	// rather than hard-coding how long a countdown mesh takes to settle.
	p := New(ModeTimeline, ModeTiming).SetTimelineLimit(4)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	c := countdown("looper", 20)
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	records := p.Timeline()
	require.NotEmpty(t, records)
	require.Greater(t, p.Cycles().Count, 4, "the run must overflow the cap to evict at all")
	assert.LessOrEqual(t, len(records), 4)

	last := records[len(records)-1]
	assert.Equal(t, p.Cycles().Count, last.Number, "the newest cycles are the ones kept")
	for i := 1; i < len(records); i++ {
		assert.Equal(t, records[i-1].Number+1, records[i].Number, "and they stay contiguous")
	}
}

func TestProfiler_TimelineRecordsWaiting(t *testing.T) {
	// Waiting is the stat a livelock shows up in, and nothing else populates it.
	p := New(ModeTimeline)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	c := testutil.MustComponent("patient",
		component.WithInputs("i1", "i2"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			if !this.InputByName("i2").HasSignals() {
				return component.ErrWaitKeepingInputs
			}
			return nil
		}))
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.Error(t, err, "the mesh stalls waiting for input that never arrives")

	records := p.Timeline()
	require.NotEmpty(t, records)
	assert.Equal(t, 1, records[0].Waiting)
}
