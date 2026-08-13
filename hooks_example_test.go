package fmesh_test

import (
	"context"
	"fmt"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// ExampleFMesh_SetupHooks observes a run from outside: a mesh-level hook counts
// activations per cycle, and a port-level hook watches signals arrive.
func ExampleFMesh_SetupHooks() {
	double, err := component.New("double",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			n, err := signal.AsInt(this.InputByName("in").Signals().First())
			if err != nil {
				return err
			}
			return this.OutputByName("out").PutPayloads(n * 2)
		}))
	if err != nil {
		panic(err)
	}

	double.InputByName("in").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(_ context.Context, added *port.SignalsAddedContext) error {
			fmt.Printf("port %q received %d signal(s)\n", added.Port.Name(), len(added.SignalsAdded))
			return nil
		})
	})

	fm, err := fmesh.New("observed")
	if err != nil {
		panic(err)
	}
	if err = fm.AddComponents(double); err != nil {
		panic(err)
	}

	fm.SetupHooks(func(h *fmesh.Hooks) {
		h.AfterCycle(func(_ context.Context, cc *fmesh.CycleContext) error {
			fmt.Printf("cycle #%d: %d activation(s)\n", cc.Cycle.Number(), cc.Cycle.ActivationResults().Len())
			return nil
		})
	})

	_ = double.InputByName("in").PutPayloads(21)

	if _, err = fm.Run(context.Background()); err != nil {
		panic(err)
	}

	result, _ := double.OutputByName("out").Signals().FirstPayload()
	fmt.Println("result:", result)
	// Output:
	// port "in" received 1 signal(s)
	// cycle #1: 1 activation(s)
	// cycle #2: 0 activation(s)
	// result: 42
}
