package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"os"
)

// This example shows that components can have external state (internal state is in experimental phase and will be added in further versions of f-mesh)
func main() {
	starter := component.New("starter").
		WithInputs("i1").
		WithOutputs("o1").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			return port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
		})

	layer1 := component.New("layer 1").
		WithDescription("This dummy bypass layer is needed to continue executing, so we will demonstrate that counter is called multiple times").
		WithInputs("i1").
		WithOutputs("o1").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			return port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
		})

	layer2 := component.New("layer 2").
		WithDescription("This dummy bypass layer is needed to continue executing, so we will demonstrate that counter is called multiple times").
		WithInputs("i1").
		WithOutputs("o1").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			return port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
		})

	layer3 := component.New("layer 3").
		WithDescription("This dummy bypass layer is needed to continue executing, so we will demonstrate that counter is called multiple times").
		WithInputs("i1").
		WithOutputs("o1").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			return port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
		})

	//This variable is not part of f-mesh and just mutated from activation function
	counterExternalState := 0

	counter := component.New("counter").
		WithDescription("Stateful counter").
		WithInputs("i1").
		WithOutputs("o1").WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
		for _, _ = range inputs.ByName("i1").AllSignalsOrNil() {
			counterExternalState++
		}

		//Latest state is always kept on o1
		outputs.ByName("o1").Clear().PutSignals(signal.New(counterExternalState))
		return nil
	})

	// Chain: starter->layer1->layer2->layer3
	starter.OutputByName("o1").PipeTo(layer1.InputByName("i1"))
	layer1.OutputByName("o1").PipeTo(layer2.InputByName("i1"))
	layer2.OutputByName("o1").PipeTo(layer3.InputByName("i1"))

	// Layers 1-3 are also reporting to the counter
	layer1.OutputByName("o1").PipeTo(counter.InputByName("i1"))
	layer2.OutputByName("o1").PipeTo(counter.InputByName("i1"))
	layer3.OutputByName("o1").PipeTo(counter.InputByName("i1"))

	fm := fmesh.New("stateful").WithComponents(starter, layer1, layer2, layer3, counter)

	//Init data (4 heterogeneous signals, value does not matter)
	starter.InputByName("i1").PutSignals(signal.NewGroup(1, "a", 0, nil).SignalsOrNil()...)

	//Run the mesh
	_, err := fm.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	count := fm.Components().ByName("counter").OutputByName("o1").FirstSignalPayloadOrDefault(0)

	//Expected: 12 (4 signals repeated 3 times (on each layer))
	fmt.Printf("Count: %d", count)

}
