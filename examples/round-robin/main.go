package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create request generator
	generator := component.New("generator").
		WithDescription("Generates requests for round-robin distribution").
		WithOutputs("requests").
		WithActivationFunc(func(this *component.Component) error {
			for i := 0; i < 12; i++ {
				msg := fmt.Sprintf("Request %d", i+1)
				this.Logger().Printf("Generating: %s", msg)
				this.OutputByName("requests").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create round-robin distributor
	distributor := component.New("distributor").
		WithDescription("Distributes requests in round-robin fashion").
		WithInputs("requests").
		WithOutputs("worker1", "worker2", "worker3").
		WithActivationFunc(func(this *component.Component) error {
			workerIndex := 0
			for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				workerIndex = (workerIndex + 1) % 3
				outputName := fmt.Sprintf("worker%d", workerIndex+1)
				this.Logger().Printf("Round-robin routing %s to %s", msg, outputName)
				this.OutputByName(outputName).PutSignals(signal.New(msg))
			}
			return nil
		})

	// Create worker components with different processing times
	workers := make([]*component.Component, 3)
	processingTimes := []time.Duration{200 * time.Millisecond, 300 * time.Millisecond, 400 * time.Millisecond}

	for i := 0; i < 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		processingTime := processingTimes[i]
		workers[i] = component.New(workerID).
			WithDescription(fmt.Sprintf("Processes requests (%s)", workerID)).
			WithInputs("requests").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s processing: %s", this.Name(), msg)
					time.Sleep(processingTime)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("round-robin-example").
		WithDescription("Demonstrates round-robin pattern using FMesh").
		WithComponents(append([]*component.Component{generator, distributor}, workers...)...)

	// Connect components
	generator.OutputByName("requests").PipeTo(distributor.InputByName("requests"))
	for i, worker := range workers {
		distributor.OutputByName(fmt.Sprintf("worker%d", i+1)).PipeTo(worker.InputByName("requests"))
	}

	// Run the mesh
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
