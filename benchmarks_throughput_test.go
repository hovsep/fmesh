package fmesh

import (
	"errors"
	"strconv"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

// The headline throughput metric for f-mesh is activation cycles per second: how
// many full run-loop ticks the scheduler can drive, with all components activating
// every cycle. It measures the library overhead added on top of the user's activation
// function, so the activation body here is deliberately near-empty.
//
// Reported custom metrics:
//   - cycles/s      — full activation cycles per second (the headline number)
//   - activations/s — component activations per second (cycles/s × mesh size)
//
// Two activation kinds isolate different parts of the overhead:
//   - dummy  — the activation returns ErrWaitingForInputsKeep, so each component keeps
//     its input and re-activates every cycle. No output ports, no pipes: this measures
//     pure scheduling + activation-lifecycle overhead (goroutine fan-out, WaitGroup,
//     result collection), with none of the drain/flush machinery.
//   - bypass — the activation copies input signals to an output port that is piped back
//     to its own input (self-loop), so one signal circulates per component forever. This
//     adds the realistic per-cycle cost: PutSignals, drain, and pipe forwarding.
//
// bypass minus dummy therefore approximates the cost of f-mesh's signal-movement path.

// cyclesPerRun is how many activation cycles each fm.Run() executes. Large enough to
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
	activationDummy activationKind = iota
	activationBypass
)

// buildThroughputMesh builds a mesh of `size` components that all activate every cycle
// and sustains that for exactly cyclesPerRun cycles per Run (time limit removed).
func buildThroughputMesh(b *testing.B, size int, kind activationKind) *FMesh {
	b.Helper()

	fm, err := New("bench-throughput",
		WithCyclesLimit(cyclesPerRun),
		WithUnlimitedTime())
	require.NoError(b, err)

	components := make([]*component.Component, size)
	for i := range size {
		name := "c" + strconv.Itoa(i)
		switch kind {
		case activationDummy:
			c, err := component.New(name,
				component.WithInputs("in"),
				component.WithActivationFunc(func(*component.Component) error {
					// Keep the input and re-activate next cycle without emitting anything.
					return component.ErrWaitingForInputsKeep
				}))
			require.NoError(b, err)
			components[i] = c
		case activationBypass:
			c, err := component.New(name,
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("out").PutSignals(this.InputByName("in").Signals().All()...)
				}))
			require.NoError(b, err)
			components[i] = c
		}
	}
	require.NoError(b, fm.AddComponents(components...))

	if kind == activationBypass {
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
		require.NoError(b, c.ClearInputs())
		require.NoError(b, c.InputByName("in").PutSignals(signal.New(0)))
	}
}

func benchmarkThroughput(b *testing.B, size int, kind activationKind) {
	fm := buildThroughputMesh(b, size, kind)

	var totalCycles int
	b.ReportAllocs()
	for b.Loop() {
		primeInputs(b, fm, size)
		ri, err := fm.Run()
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

// BenchmarkMeshThroughputDummy measures cycles/s with a near-empty activation function
// (no signal movement) — the scheduling/activation-lifecycle overhead floor.
func BenchmarkMeshThroughputDummy(b *testing.B) {
	for _, s := range throughputSizes {
		b.Run(s.name, func(b *testing.B) { benchmarkThroughput(b, s.size, activationDummy) })
	}
}

// BenchmarkMeshThroughputBypass measures cycles/s with a passthrough activation and a
// self-loop, exercising the full per-cycle drain/flush/pipe path.
func BenchmarkMeshThroughputBypass(b *testing.B) {
	for _, s := range throughputSizes {
		b.Run(s.name, func(b *testing.B) { benchmarkThroughput(b, s.size, activationBypass) })
	}
}
