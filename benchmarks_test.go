package fmesh

import (
	"strconv"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

// buildPipelineMesh builds a linear pipeline: c0 -> c1 -> ... -> c(n-1).
// Each component increments the incoming integer payload by 1.
func buildPipelineMesh(b *testing.B, componentsCount int) *FMesh {
	b.Helper()

	fm, err := New("bench-pipeline")
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

	for i := 0; i < componentsCount-1; i++ {
		require.NoError(b, components[i].OutputByName("out").PipeTo(components[i+1].InputByName("in")))
	}

	return fm
}

// BenchmarkMeshRunPipeline measures full mesh execution over a linear
// 10-component pipeline (activation + drain path).
func BenchmarkMeshRunPipeline(b *testing.B) {
	fm := buildPipelineMesh(b, 10)
	firstInput := fm.ComponentByName("c0").InputByName("in")

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := firstInput.PutSignals(signal.New(0)); err != nil {
			b.Fatal(err)
		}
		if _, err := fm.Run(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMeshRunFanOut measures mesh execution with one producer fanning out
// to 10 consumers (flush/forward path).
func BenchmarkMeshRunFanOut(b *testing.B) {
	const consumersCount = 10

	fm, err := New("bench-fanout")
	require.NoError(b, err)

	producer, err := component.New("producer",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(this *component.Component) error {
			return this.OutputByName("out").PutSignals(this.InputByName("in").Signals().All()...)
		}))
	require.NoError(b, err)
	require.NoError(b, fm.AddComponents(producer))

	for i := range consumersCount {
		consumer, err := component.New("consumer"+strconv.Itoa(i),
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(this *component.Component) error {
				num := this.InputByName("in").Signals().FirstPayloadOrDefault(0).(int)
				return this.OutputByName("out").PutSignals(signal.New(num * 2))
			}))
		require.NoError(b, err)
		require.NoError(b, fm.AddComponents(consumer))
		require.NoError(b, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	}

	producerInput := producer.InputByName("in")

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := producerInput.PutSignals(signal.New(42)); err != nil {
			b.Fatal(err)
		}
		if _, err := fm.Run(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMeshConstruction measures building a 10-component mesh with ports
// and pipes (mesh manipulation path).
func BenchmarkMeshConstruction(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		buildPipelineMesh(b, 10)
	}
}
