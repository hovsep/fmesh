package profiler

import (
	"context"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiler_Pipes(t *testing.T) {
	p := New(ModeThroughput)

	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	require.NoError(t, fm.AddComponents(
		testutil.MustComponent("producer",
			component.WithInputs("i1"),
			component.WithOutputs("o1", "debug"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				return this.OutputByName("o1").PutPayloads(1, 2, 3)
			})),
		testutil.MustComponent("consumer",
			component.WithInputs("i1"),
			component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
		testutil.MustComponent("logger",
			component.WithInputs("i1"),
			component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
	))

	producer := fm.Components().ByName("producer")
	// Wired after AddComponents, which is the case a plugin that instrumented
	// pipes rather than ports would miss.
	require.NoError(t, producer.OutputByName("o1").
		PipeTo(fm.Components().ByName("consumer").InputByName("i1")))
	require.NoError(t, producer.OutputByName("debug").
		PipeTo(fm.Components().ByName("logger").InputByName("i1")))
	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	pipes := p.Pipes()
	require.Len(t, pipes, 2)

	assert.Equal(t, "producer.o1", pipes[0].Source)
	assert.Equal(t, "consumer.i1", pipes[0].Destination)
	assert.Equal(t, "producer.o1 -> consumer.i1", pipes[0].Pipe())
	assert.Equal(t, 1, pipes[0].Transfers)
	assert.Equal(t, 3, pipes[0].Signals)
	assert.Equal(t, 3, pipes[0].Min)
	assert.Equal(t, 3, pipes[0].Max)
	assert.InDelta(t, 3.0, pipes[0].Avg(), 1e-9)

	// A pipe nothing ever used is the coldest pipe there is, and it can only be
	// reported because pipes are registered as well as counted.
	assert.Equal(t, "producer.debug -> logger.i1", pipes[1].Pipe())
	assert.Zero(t, pipes[1].Transfers)
	assert.Zero(t, pipes[1].Avg())

	assert.Contains(t, p.Report(), "producer.o1 -> consumer.i1")

	t.Run("Reset zeroes the flows but keeps the topology", func(t *testing.T) {
		p.Reset()

		after := p.Pipes()
		require.Len(t, after, 2, "both pipes still exist, they have just carried nothing")
		for _, s := range after {
			assert.Zero(t, s.Transfers)
			assert.Zero(t, s.Signals)
		}
	})
}

func TestProfiler_PipesRankHotAndCold(t *testing.T) {
	// Pipes() is volume-sorted so the hottest is first and the coldest last;
	// TopNPipes ranks by how often instead, which is a different question.
	p := New(ModeThroughput)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	looper := testutil.MustComponent("looper",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			left, _ := this.State().GetOrDefault("left", 4).(int)
			if left <= 1 {
				return nil
			}
			this.State().Set("left", left-1)
			return this.OutputByName("o1").PutPayloads(left - 1)
		}))
	require.NoError(t, looper.LoopbackPipe("o1", "i1"))

	bulk := testutil.MustComponent("bulk",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("o1").PutPayloads(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
		}))
	sink := testutil.MustComponent("sink",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(looper, bulk, sink))
	require.NoError(t, bulk.OutputByName("o1").PipeTo(sink.InputByName("i1")))
	require.NoError(t, looper.InputByName("i1").PutSignals(signal.New("go")))
	require.NoError(t, bulk.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	byVolume := p.Pipes()
	require.Len(t, byVolume, 2)
	assert.Equal(t, "bulk.o1 -> sink.i1", byVolume[0].Pipe(), "most signals first")
	assert.Equal(t, 10, byVolume[0].Signals)

	byFrequency := p.TopNPipes(2)
	require.Len(t, byFrequency, 2)
	assert.Equal(t, "looper.o1 -> looper.i1", byFrequency[0].Pipe(), "most transfers first")
	assert.Greater(t, byFrequency[0].Transfers, byFrequency[1].Transfers)

	assert.Empty(t, p.TopNPipes(0))
	assert.Empty(t, p.TopNPipes(-1), "a negative n returns nothing rather than panicking")
	assert.Len(t, p.TopNPipes(100), 2, "asking for more than exists returns everything")
}

func TestProfiler_PipeBatchSizes(t *testing.T) {
	// Nothing else exercises a Min and Max that differ: a component emitting a
	// different number of signals on each of two cycles.
	p := New(ModeThroughput)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	producer := testutil.MustComponent("producer",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			// One signal on the first cycle, three on the second, then stop.
			sent, _ := this.State().GetOrDefault("sent", 0).(int)
			this.State().Set("sent", sent+1)
			switch sent {
			case 0:
				return this.OutputByName("o1").PutPayloads(1)
			case 1:
				return this.OutputByName("o1").PutPayloads(1, 2, 3)
			default:
				return nil
			}
		}))
	require.NoError(t, producer.LoopbackPipe("o1", "i1"))

	require.NoError(t, fm.AddComponents(producer))
	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	pipes := p.Pipes()
	require.Len(t, pipes, 1)
	assert.Equal(t, 1, pipes[0].Min, "the first cycle sent one signal")
	assert.Equal(t, 3, pipes[0].Max, "the second sent three")
}

func TestProfiler_PipeTiesBreakOnLabel(t *testing.T) {
	// Equal traffic must not reorder a report between runs. Detached ports also
	// pin the nil-parent label: "out", not "<nil>.out".
	p := New(ModeThroughput)

	charlie, alpha, bravo := mustPipeKey("charlie"), mustPipeKey("alpha"), mustPipeKey("bravo")
	p.pipes = map[pipeKey]Flow{
		charlie: {Transfers: 2, Signals: 5},
		alpha:   {Transfers: 2, Signals: 5},
		bravo:   {Transfers: 9, Signals: 1},
	}

	byVolume := make([]string, 0, 3)
	for _, s := range p.Pipes() {
		byVolume = append(byVolume, s.Source)
	}
	assert.Equal(t, []string{"alpha_out", "charlie_out", "bravo_out"}, byVolume,
		"most signals first, equal volumes in label order")

	byFrequency := make([]string, 0, 3)
	for _, s := range p.TopNPipes(3) {
		byFrequency = append(byFrequency, s.Source)
	}
	assert.Equal(t, []string{"bravo_out", "alpha_out", "charlie_out"}, byFrequency,
		"most transfers first, equal counts in label order")
}

func TestProfiler_PortsAddedAfterArrivalAreInvisible(t *testing.T) {
	// A mesh plugin only sees a component when it arrives, so ports added later
	// carry no hook. Documented rather than fixed -- pinned here so it stays a
	// decision.
	p := New(ModeThroughput)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	producer := testutil.MustComponent("producer",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("late").PutPayloads(1)
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	require.NoError(t, fm.AddComponents(producer, consumer))

	require.NoError(t, producer.AddOutputs("late"))
	require.NoError(t, producer.OutputByName("late").PipeTo(consumer.InputByName("i1")))
	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	assert.Empty(t, p.Pipes(), "a port the plugin never saw carries no hook")
}

// mustPipeKey builds a detached pipe whose ends are labeled by port name alone.
func mustPipeKey(name string) pipeKey {
	out, err := port.NewOutput(name + "_out")
	if err != nil {
		panic(err)
	}
	in, err := port.NewInput(name + "_in")
	if err != nil {
		panic(err)
	}
	return pipeKey{source: out, destination: in}
}
