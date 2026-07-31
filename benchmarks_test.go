package fmesh

import (
	"context"
	"strconv"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

// What a run costs (BenchmarkMeshRun) and what building a mesh costs
// (BenchmarkMeshConstruction). Throughput lives in benchmarks_throughput_test.go.
//
// Sizes are swept so the complexity class shows up in benchstat output: a curve
// that bends is worth catching, an absolute number on a shared runner is not.

// benchSizes is the standard sweep. See .agent/docs/benchmarking.md.
var benchSizes = []int{10, 100, 1_000, 10_000}

// incrementer is the standard benchmark component: read an int, emit it plus one.
// Deliberately trivial, so the numbers measure f-mesh rather than user code.
func incrementer(b *testing.B, name string) *component.Component {
	b.Helper()
	c, err := component.New(name,
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			num := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
			return this.OutputByName("out").PutSignals(signal.New(num + 1))
		}))
	require.NoError(b, err)
	return c
}

// buildWideMesh: n independent components, no pipes. Seeding every input activates
// all of them in one cycle, isolating the run loop's one-goroutine-per-component
// cost.
//
// Wide, not deep: a linear pipeline of n components needs n cycles and would hit
// the default limits, which is why pipeline is benchmarked at one fixed size.
func buildWideMesh(b *testing.B, n int) *FMesh {
	b.Helper()

	fm, err := New("bench-wide")
	require.NoError(b, err)

	components := make([]*component.Component, n)
	for i := range n {
		components[i] = incrementer(b, "c"+strconv.Itoa(i))
	}
	require.NoError(b, fm.AddComponents(components...))
	return fm
}

// buildFanInMesh: n sources all piping into one collector input.
func buildFanInMesh(b *testing.B, n int) *FMesh {
	b.Helper()

	fm, err := New("bench-fan-in")
	require.NoError(b, err)

	collector, err := component.New("collector",
		component.WithInputs("in"),
		component.WithActivationFunc(func(_ context.Context, _ *component.Component) error {
			return nil
		}))
	require.NoError(b, err)

	sources := make([]*component.Component, n)
	for i := range n {
		sources[i] = incrementer(b, "c"+strconv.Itoa(i))
	}
	require.NoError(b, fm.AddComponents(sources...))
	require.NoError(b, fm.AddComponents(collector))

	for _, source := range sources {
		require.NoError(b, source.OutputByName("out").PipeTo(collector.InputByName("in")))
	}
	return fm
}

// buildFanOutMesh: one producer piping into n consumers.
func buildFanOutMesh(b *testing.B, n int) *FMesh {
	b.Helper()

	fm, err := New("bench-fan-out")
	require.NoError(b, err)

	producer := incrementer(b, "c0")
	require.NoError(b, fm.AddComponents(producer))

	for i := range n {
		consumer := incrementer(b, "consumer"+strconv.Itoa(i))
		require.NoError(b, fm.AddComponents(consumer))
		require.NoError(b, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	}
	return fm
}

// buildPipelineMesh: a linear chain c0 -> c1 -> ... -> c(n-1).
func buildPipelineMesh(b *testing.B, n int) *FMesh {
	b.Helper()

	fm, err := New("bench-pipeline")
	require.NoError(b, err)

	components := make([]*component.Component, n)
	for i := range n {
		components[i] = incrementer(b, "c"+strconv.Itoa(i))
	}
	require.NoError(b, fm.AddComponents(components...))

	for i := range n - 1 {
		require.NoError(b, components[i].OutputByName("out").PipeTo(components[i+1].InputByName("in")))
	}
	return fm
}

// seedAll puts one signal on the "in" port of every c<i> component, so they all
// activate in the next cycle.
func seedAll(b *testing.B, fm *FMesh, n int) {
	b.Helper()
	for i := range n {
		in := fm.ComponentByName("c" + strconv.Itoa(i)).InputByName("in")
		if err := in.PutSignals(signal.New(0)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMeshRun measures a full run — activation plus drain — across four
// topologies: wide (goroutine fan-out), fan-in (per-signal appends into one
// port), fan-out (pipe forwarding) and pipeline (n cycles of the whole loop).
//
// fan-in is the tripwire: its curve is quadratic today and would flatten to
// linear if pipe forwarding ever batched its appends.
func BenchmarkMeshRun(b *testing.B) {
	topologies := []struct {
		name  string
		sizes []int
		build func(*testing.B, int) *FMesh
	}{
		{"wide", benchSizes, buildWideMesh},
		{"fan-in", benchSizes, buildFanInMesh},
		{"fan-out", benchSizes, buildFanOutMesh},
		{"pipeline", []int{10}, buildPipelineMesh},
	}

	for _, topology := range topologies {
		b.Run(topology.name, func(b *testing.B) {
			for _, n := range topology.sizes {
				b.Run(strconv.Itoa(n), func(b *testing.B) {
					fm := topology.build(b, n)
					// A pipeline is seeded at its head only; the rest are fed everywhere.
					seeded := n
					if topology.name != "wide" && topology.name != "fan-in" {
						seeded = 1
					}
					b.ReportAllocs()
					for b.Loop() {
						seedAll(b, fm, seeded)
						if _, err := fm.Run(context.Background()); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

// BenchmarkMeshConstruction measures building a linear mesh of n components and
// pipes. The sweep should stay linear.
func BenchmarkMeshConstruction(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				buildPipelineMesh(b, n)
			}
		})
	}
}
