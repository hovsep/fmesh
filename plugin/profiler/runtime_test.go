package profiler

import (
	"context"
	"testing"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allocating deliberately allocates so a resource delta has something to report.
func allocating(name string) *component.Component {
	return testutil.MustComponent(name,
		component.WithInputs("i1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			junk := make([][]byte, 0, 512)
			for range 512 {
				junk = append(junk, make([]byte, 1024))
			}
			this.State().Set("kept", len(junk))
			return nil
		}))
}

func TestProfiler_Resources(t *testing.T) {
	p := New(ModeRuntime)
	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	c := allocating("hungry")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	r := p.Resources()
	assert.Positive(t, r.AllocBytes, "half a megabyte of activation work is visible")
	assert.Positive(t, r.AllocObjects)
	// CPU is only advanced by the runtime at GC mark termination, so a run that
	// completed no GC cycle reports zero however long it took. Nothing stronger
	// than non-negative can be asserted.
	assert.GreaterOrEqual(t, r.CPUUser, time.Duration(0))
	assert.GreaterOrEqual(t, r.CPUGC, time.Duration(0))
	assert.Contains(t, p.Report(), "process-wide deltas")

	t.Run("Reset discards the deltas", func(t *testing.T) {
		p.Reset()
		assert.Zero(t, p.Resources().AllocBytes)
	})
}

func TestProfiler_ResourcesPerCycleNeedATimeline(t *testing.T) {
	// The cycle record is the only place a per-cycle resource number can live, so
	// runtime sampling at cycle boundaries is implied by the two modes together
	// rather than being a mode of its own.
	withTimeline := New(ModeRuntime, ModeTimeline)
	fm, err := fmesh.New("m", fmesh.WithPlugins(withTimeline))
	require.NoError(t, err)

	c := allocating("hungry")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	records := withTimeline.Timeline()
	require.NotEmpty(t, records)
	assert.Positive(t, records[0].Resources.AllocBytes,
		"the cycle that did the allocating carries the delta")

	runtimeOnly := New(ModeRuntime)
	assert.Empty(t, runtimeOnly.Timeline(), "no timeline, nowhere to put per-cycle numbers")
}

func TestProfiler_ModesGateHookRegistration(t *testing.T) {
	// A dimension that is off registers no hooks at all, so nothing it would
	// have measured appears -- and the default profiler is unchanged.
	p := New()
	require.Equal(t, ModeTiming, p.Modes())

	fm, err := fmesh.New("m", fmesh.WithPlugins(p))
	require.NoError(t, err)

	producer := testutil.MustComponent("producer",
		component.WithInputs("i1"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("o1").PutPayloads(1)
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("i1"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("o1").PipeTo(consumer.InputByName("i1")))
	require.NoError(t, producer.InputByName("i1").PutSignals(signal.New("go")))

	_, err = fm.Run(context.Background())
	require.NoError(t, err)

	assert.Positive(t, p.Runs().Count, "timing is on")
	assert.Empty(t, p.Pipes())
	assert.Empty(t, p.Timeline())
	assert.Zero(t, p.Resources())

	report := p.Report()
	assert.NotContains(t, report, "modes:", "a default profiler reports what it always did")
	assert.NotContains(t, report, "pipe ")
	assert.NotContains(t, report, "timeline:")
	assert.NotContains(t, report, "resources ")
}

func TestProfileMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
		want string
	}{
		{name: "no modes at all", mode: 0, want: "none"},
		{name: "one", mode: ModeTiming, want: "timing"},
		{
			name: "combined renders in bit order, not argument order",
			mode: ModeRuntime | ModeTiming,
			want: "timing|runtime",
		},
		{name: "all", mode: ModeAll, want: "timing|throughput|timeline|runtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.mode.String())
		})
	}
}

func TestNewProfiler_Modes(t *testing.T) {
	assert.Equal(t, ModeTiming, New().Modes(), "no arguments is what it always was")
	assert.Equal(t, ModeTiming, New(0).Modes(), "an empty mode set is not a silent no-op")
	assert.Equal(t, ModeAll, New(ModeAll).Modes())
	assert.Equal(t,
		New(ModeTiming|ModeRuntime).Modes(),
		New(ModeTiming, ModeRuntime).Modes(),
		"OR'd together or passed separately, the same thing")
}
