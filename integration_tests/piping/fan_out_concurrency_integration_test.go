package piping

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFanOut_threeConsumers_seeSameSignalPointer asserts fan-out wiring: one
// signal instance is delivered to every downstream port, and fmesh activates
// components in parallel (see fmesh.runCycle).
func TestFanOut_threeConsumers_seeSameSignalPointer(t *testing.T) {
	var ptrs sync.Map

	producer := mustComponent("producer",
		component.WithInputs("start"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(this *component.Component) error {
			return this.OutputByName("o1").PutSignals(signal.New(42).WithLabel("route", "fan"))
		}))

	makeConsumer := func(name, slot string) *component.Component {
		return mustComponent(name,
			component.WithInputs("i1"),
			component.WithOutputs("o1"),
			component.WithActivationFunc(func(this *component.Component) error {
				first := this.InputByName("i1").Signals().First()
				if first != nil {
					ptrs.Store(slot, uintptr(unsafe.Pointer(first)))
				}
				return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
			}))
	}

	c1 := makeConsumer("consumer1", "1")
	c2 := makeConsumer("consumer2", "2")
	c3 := makeConsumer("consumer3", "3")

	fm := mustFMesh("fan-out-same-pointer")
	require.NoError(t, fm.AddComponents(producer, c1, c2, c3))
	require.NoError(t, fm.Components().ByName("producer").OutputByName("o1").PipeTo(
		fm.Components().ByName("consumer1").InputByName("i1"),
		fm.Components().ByName("consumer2").InputByName("i1"),
		fm.Components().ByName("consumer3").InputByName("i1"),
	))

	require.NoError(t, fm.Components().ByName("producer").InputByName("start").PutSignals(signal.New(struct{}{})))

	_, err := fm.Run()
	require.NoError(t, err)

	p1, ok1 := ptrs.Load("1")
	p2, ok2 := ptrs.Load("2")
	p3, ok3 := ptrs.Load("3")
	require.True(t, ok1 && ok2 && ok3, "each consumer should have recorded a non-nil *signal.Signal")

	assert.Equal(t, p1, p2, "fan-out must deliver the same *Signal pointer to each consumer (#203)")
	assert.Equal(t, p2, p3, "fan-out must deliver the same *Signal pointer to each consumer (#203)")
}

// TestFanOut_sharedSignal_parallelStress_completes exercises the same fan-out
// plus concurrent label work on the shared *signal.Signal. With copy-on-write
// signals this completes cleanly; run with -race to confirm no data races.
func TestFanOut_sharedSignal_parallelStress_completes(t *testing.T) {
	const stressIters = 400

	var ptrs sync.Map

	producer := mustComponent("producer",
		component.WithInputs("start"),
		component.WithOutputs("o1"),
		component.WithActivationFunc(func(this *component.Component) error {
			return this.OutputByName("o1").PutSignals(signal.New(1).WithLabel("seed", "x"))
		}))

	makeConsumer := func(name string, mode int) *component.Component {
		return mustComponent(name,
			component.WithInputs("i1"),
			component.WithOutputs("o1"),
			component.WithActivationFunc(func(this *component.Component) error {
				shared := this.InputByName("i1").Signals().First()
				if shared == nil {
					return nil
				}
				ptrs.Store(name, uintptr(unsafe.Pointer(shared)))

				switch mode {
				case 0:
					for i := range stressIters {
						_ = shared.MapPayload(func(p any) any {
							_ = i
							return p
						})
					}
				case 1:
					for i := range stressIters {
						shared.WithLabel(fmt.Sprintf("w1_%d", i), "v")
					}
				default:
					for i := range stressIters {
						shared.WithLabel(fmt.Sprintf("w2_%d", i), "v")
					}
				}
				return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
			}))
	}

	fm, err := fmesh.New("fan-out-stress", fmesh.WithConfig(fmesh.Config{
		ErrorHandlingStrategy: fmesh.IgnoreAll,
	}))
	require.NoError(t, err)
	require.NoError(t, fm.AddComponents(
		producer,
		makeConsumer("consumerA", 0),
		makeConsumer("consumerB", 1),
		makeConsumer("consumerC", 2),
	))

	require.NoError(t, fm.Components().ByName("producer").OutputByName("o1").PipeTo(
		fm.Components().ByName("consumerA").InputByName("i1"),
		fm.Components().ByName("consumerB").InputByName("i1"),
		fm.Components().ByName("consumerC").InputByName("i1"),
	))

	require.NoError(t, fm.Components().ByName("producer").InputByName("start").PutSignals(signal.New(struct{}{})))

	_, err = fm.Run()
	require.NoError(t, err)

	pA, okA := ptrs.Load("consumerA")
	pB, okB := ptrs.Load("consumerB")
	pC, okC := ptrs.Load("consumerC")
	require.True(t, okA && okB && okC)

	assert.Equal(t, pA, pB)
	assert.Equal(t, pB, pC)
}
