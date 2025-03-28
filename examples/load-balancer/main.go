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
		WithDescription("Generates requests to be load balanced").
		WithOutputs("requests").
		WithActivationFunc(func(this *component.Component) error {
			for i := 0; i < 10; i++ {
				msg := fmt.Sprintf("Request %d", i+1)
				this.Logger().Printf("Generating: %s", msg)
				this.OutputByName("requests").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create load balancer
	loadBalancer := component.New("load-balancer").
		WithDescription("Distributes requests across workers").
		WithInputs("requests").
		WithOutputs("worker1", "worker2", "worker3").
		WithActivationFunc(func(this *component.Component) error {
			workerIndex := 0
			for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				workerIndex = (workerIndex + 1) % 3
				outputName := fmt.Sprintf("worker%d", workerIndex+1)
				this.Logger().Printf("Routing %s to %s", msg, outputName)
				this.OutputByName(outputName).PutSignals(signal.New(msg))
			}
			return nil
		})

	// Create worker components
	workers := make([]*component.Component, 3)
	for i := 0; i < 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		workers[i] = component.New(workerID).
			WithDescription(fmt.Sprintf("Processes requests (%s)", workerID)).
			WithInputs("requests").
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s processing: %s", this.Name(), msg)
					time.Sleep(200 * time.Millisecond)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("load-balancer-example").
		WithDescription("Demonstrates load balancer pattern using FMesh").
		WithComponents(append([]*component.Component{generator, loadBalancer}, workers...)...)

	// Connect components
	generator.OutputByName("requests").PipeTo(loadBalancer.InputByName("requests"))
	for i, worker := range workers {
		loadBalancer.OutputByName(fmt.Sprintf("worker%d", i+1)).PipeTo(worker.InputByName("requests"))
	}

	// Run the mesh
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
