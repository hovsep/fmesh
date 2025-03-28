package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create data generator
	generator := component.New("generator").
		WithDescription("Generates data for pipeline processing").
		WithOutputs("data").
		WithActivationFunc(func(this *component.Component) error {
			for i := 0; i < 5; i++ {
				msg := fmt.Sprintf("Data %d", i+1)
				this.Logger().Printf("Generating: %s", msg)
				this.OutputByName("data").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create stage 1: Data preparation
	stage1 := component.New("stage1").
		WithDescription("Prepares data for processing").
		WithInputs("data").
		WithOutputs("prepared").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Stage 1 preparing: %s", msg)
				time.Sleep(200 * time.Millisecond)
				this.OutputByName("prepared").PutSignals(signal.New(msg))
			}
			return nil
		})

	// Create stage 2: Data processing
	stage2 := component.New("stage2").
		WithDescription("Processes prepared data").
		WithInputs("data").
		WithOutputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Stage 2 processing: %s", msg)
				time.Sleep(300 * time.Millisecond)
				this.OutputByName("processed").PutSignals(signal.New(msg))
			}
			return nil
		})

	// Create stage 3: Data transformation
	stage3 := component.New("stage3").
		WithDescription("Transforms processed data").
		WithInputs("data").
		WithOutputs("transformed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Stage 3 transforming: %s", msg)
				time.Sleep(200 * time.Millisecond)
				this.OutputByName("transformed").PutSignals(signal.New(msg))
			}
			return nil
		})

	// Create final stage: Data collection
	collector := component.New("collector").
		WithDescription("Collects transformed data").
		WithInputs("data").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("data").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(string)
				this.Logger().Printf("Collecting: %s", msg)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("pipeline-example").
		WithDescription("Demonstrates pipeline pattern using FMesh").
		WithComponents(generator, stage1, stage2, stage3, collector)

	// Connect components in pipeline
	generator.OutputByName("data").PipeTo(stage1.InputByName("data"))
	stage1.OutputByName("prepared").PipeTo(stage2.InputByName("data"))
	stage2.OutputByName("processed").PipeTo(stage3.InputByName("data"))
	stage3.OutputByName("transformed").PipeTo(collector.InputByName("data"))

	// Run the mesh
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
