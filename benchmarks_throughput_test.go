package fmesh

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

// Throughput in activation cycles per second, reported as cycles/s and
// activations/s. The activation body is near-empty so the numbers measure f-mesh
// rather than user code.
//
// Two kinds bracket the cost of moving signals: scheduling-only keeps its input
// and emits nothing (goroutine fan-out and result collection only), while
// with-signal-movement circulates one signal per component through a self-loop.
// The difference between them is the price of the signal-movement path.

// cyclesPerRun is how many activation cycles each fm.Run(context.Background()) executes. Large enough to
// amortize per-Run setup so the measured rate reflects steady-state per-cycle overhead.
const cyclesPerRun = 200

// throughputSizes covers the mesh scales called out in the design: tens, hundreds,
// thousands of components.
var throughputSizes = []struct {
	name string
	size int
}{
	{"small", 50},
	{"mid", 500},
	{"huge", 5000},
}

type activationKind int

const (
	activationSchedulingOnly activationKind = iota
	activationSignalMovement
)

// buildThroughputMesh builds a mesh of `size` components that all activate every cycle
// and sustains that for exactly cyclesPerRun cycles per Run (time limit removed).
func buildThroughputMesh(b *testing.B, size int, kind activationKind) *FMesh {
	b.Helper()

	fm, err := New("bench-throughput",
		WithCyclesLimit(cyclesPerRun),
		WithUnlimitedTime(),
		// scheduling-only components wait forever on purpose — that is the point
		// of the measurement, and exactly what the livelock detector exists to stop.
		WithoutLivelockDetection())
	require.NoError(b, err)

	components := make([]*component.Component, size)
	for i := range size {
		name := "c" + strconv.Itoa(i)
		switch kind {
		case activationSchedulingOnly:
			c, err := component.New(name,
				component.WithInputs("in"),
				component.WithActivationFunc(func(context.Context, *component.Component) error {
					// Keep the input and re-activate next cycle without emitting anything.
					return component.ErrWaitKeepingInputs
				}))
			require.NoError(b, err)
			components[i] = c
		case activationSignalMovement:
			c, err := component.New(name,
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					return this.OutputByName("out").PutSignals(this.InputByName("in").Signals().All()...)
				}))
			require.NoError(b, err)
			components[i] = c
		}
	}
	require.NoError(b, fm.AddComponents(components...))

	if kind == activationSignalMovement {
		// Self-loop: each output feeds its own input so a signal circulates forever.
		for _, c := range components {
			require.NoError(b, c.OutputByName("out").PipeTo(c.InputByName("in")))
		}
	}

	return fm
}

// primeInputs puts exactly one signal on every component's input, clearing first so no
// signal accumulates across Runs (the "keep" kind retains inputs between Runs).
func primeInputs(b *testing.B, fm *FMesh, size int) {
	b.Helper()
	for i := range size {
		c := fm.ComponentByName("c" + strconv.Itoa(i))
		require.NoError(b, c.ClearInputs(context.Background()))
		require.NoError(b, c.InputByName("in").PutSignals(signal.New(0)))
	}
}

func benchmarkThroughput(b *testing.B, size int, kind activationKind) {
	fm := buildThroughputMesh(b, size, kind)

	var totalCycles int
	b.ReportAllocs()
	for b.Loop() {
		primeInputs(b, fm, size)
		ri, err := fm.Run(context.Background())
		// Hitting the cycle limit is the expected, normal stop for a sustained mesh.
		if err != nil && !errors.Is(err, ErrReachedMaxAllowedCycles) {
			b.Fatal(err)
		}
		totalCycles += ri.Cycles.Len()
	}

	secs := b.Elapsed().Seconds()
	b.ReportMetric(float64(totalCycles)/secs, "cycles/s")
	b.ReportMetric(float64(totalCycles*size)/secs, "activations/s")
}

// BenchmarkMeshCyclesPerSecond reports how fast the run loop ticks, with and
// without the signal-movement path.
func BenchmarkMeshCyclesPerSecond(b *testing.B) {
	kinds := []struct {
		name string
		kind activationKind
	}{
		{"scheduling-only", activationSchedulingOnly},
		{"with-signal-movement", activationSignalMovement},
	}

	for _, k := range kinds {
		b.Run(k.name, func(b *testing.B) {
			for _, size := range throughputSizes {
				b.Run(size.name, func(b *testing.B) { benchmarkThroughput(b, size.size, k.kind) })
			}
		})
	}
}
