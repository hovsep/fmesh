package fmesh

import (
	"strconv"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

// benchSizes is the standard size sweep for scale benchmarks. Sweeping the size
// (rather than fixing one) makes the complexity class visible in benchstat output.
var benchSizes = []int{10, 100, 1_000, 10_000}

// buildWideMesh builds a mesh of componentsCount independent components (no pipes
// between them). Seeding every input and calling Run activates all of them in a
// single cycle, so this isolates the run loop's one-goroutine-per-component cost.
//
// Wide, not deep: a linear pipeline of N components would need N cycles and hit the
// default CyclesLimit (1000) / TimeLimit (5s). A wide mesh runs in one activation cycle.
func buildWideMesh(b *testing.B, componentsCount int) *FMesh {
	b.Helper()

	fm, err := New("bench-wide")
	require.NoError(b, err)

	components := make([]*component.Component, componentsCount)
	for i := range componentsCount {
		c, err := component.New("c"+strconv.Itoa(i),
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(this *component.Component) error {
				num := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
				return this.OutputByName("out").PutSignals(signal.New(num + 1))
			}))
		require.NoError(b, err)
		components[i] = c
	}
	require.NoError(b, fm.AddComponents(components...))

	return fm
}

// seedWideMesh puts one signal on every component's input so all activate next cycle.
func seedWideMesh(b *testing.B, fm *FMesh, componentsCount int) {
	b.Helper()
	for i := range componentsCount {
		in := fm.ComponentByName("c" + strconv.Itoa(i)).InputByName("in")
		if err := in.PutSignals(signal.New(0)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMeshRunWide measures full mesh execution when N components all activate in
// the same cycle — i.e. N concurrent goroutines spawned by the run loop. Sweeping N
// shows how activation overhead scales with mesh width.
func BenchmarkMeshRunWide(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			fm := buildWideMesh(b, n)
			b.ReportAllocs()
			for b.Loop() {
				seedWideMesh(b, fm, n)
				if _, err := fm.Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// buildFanInMesh builds componentsCount source components whose outputs all pipe into a
// single collector input. Draining cycle 1 forwards one signal per source into the same
// destination port, and each append copies the collector's whole signal group
// (port.putSignals → signal.Group.With), so the drain is O(N²) in the number of sources.
func buildFanInMesh(b *testing.B, componentsCount int) *FMesh {
	b.Helper()

	fm, err := New("bench-fan-in")
	require.NoError(b, err)

	collector, err := component.New("collector",
		component.WithInputs("in"),
		component.WithActivationFunc(func(*component.Component) error {
			// Consume silently: the measured cost is the fan-in drain, not the collector.
			return nil
		}))
	require.NoError(b, err)

	sources := make([]*component.Component, componentsCount)
	for i := range componentsCount {
		c, err := component.New("c"+strconv.Itoa(i),
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(this *component.Component) error {
				num := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
				return this.OutputByName("out").PutSignals(signal.New(num + 1))
			}))
		require.NoError(b, err)
		require.NoError(b, c.OutputByName("out").PipeTo(collector.InputByName("in")))
		sources[i] = c
	}
	require.NoError(b, fm.AddComponents(sources...))
	require.NoError(b, fm.AddComponents(collector))

	return fm
}

// BenchmarkMeshFanIn measures full mesh execution when N sources all pipe into one
// collector input — the fan-in topology. Sweeping N is the tripwire for the drain's
// per-signal-append group copying: today the curve is quadratic; if pipe forwarding
// ever batches appends into the destination port, it flattens to linear.
func BenchmarkMeshFanIn(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			fm := buildFanInMesh(b, n)
			b.ReportAllocs()
			for b.Loop() {
				seedWideMesh(b, fm, n)
				if _, err := fm.Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMeshConstructionScale measures building a linear mesh of N components and
// pipes. Sweeping N should show construction staying linear.
func BenchmarkMeshConstructionScale(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				buildPipelineMesh(b, n)
			}
		})
	}
}
