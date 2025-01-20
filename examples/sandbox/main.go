package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"math/rand"
	"os"
)

func main() {

	generator := component.New("generator").
		WithDescription("generates a random number").
		WithInputs("trigger").
		WithOutputs("number").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			outputs.ByName("number").PutSignals(signal.New(rand.Int()))
			return nil
		})

	counter := component.New("counter").
		WithDescription("counts the number of observed signals on i1").
		WithInputs("signals", "state_in").
		WithOutputs("count", "state_out").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
			oldCount := inputs.ByName("state_in").FirstSignalPayloadOrDefault(0).(int)
			increment := inputs.ByName("signals").Buffer().Len()
			newCount := oldCount + increment

			if newCount > oldCount {
				outputs.ByName("count").PutSignals(signal.New(newCount))
			}
			outputs.ByName("state_out").PutSignals(signal.New(newCount))

			return nil
		})

	limiter := component.New("limiter").
		WithDescription("decides when to stop generating numbers").
		WithInputs("count").
		WithOutputs("trigger_gen").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {

			c := inputs.ByName("count").FirstSignalPayloadOrDefault(0).(int)

			if c < 3 {
				outputs.ByName("trigger_gen").PutSignals(signal.New("more numbers"))
			}

			return nil
		})

	generator.OutputByName("number").PipeTo(counter.InputByName("signals"))
	limiter.OutputByName("trigger_gen").PipeTo(generator.InputByName("trigger"))
	counter.OutputByName("state_out").PipeTo(counter.InputByName("state_in"))
	counter.OutputByName("count").PipeTo(limiter.InputByName("count"))

	fm := fmesh.New("test").WithComponents(generator, counter, limiter)

	fm.Components().ByName("generator").InputByName("trigger").PutSignals(signal.New("fire"))

	_, err := fm.Run()
	if err != nil {
		fmt.Println("MESH returned an error:", err)
		os.Exit(1)
	}

	c := fm.Components().ByName("counter").OutputByName("count").FirstSignalPayloadOrNil().(int)
	fmt.Println("Counter state :", c)
}
