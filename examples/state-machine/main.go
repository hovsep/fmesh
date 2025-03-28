package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// State Machine Pattern Example
// This example demonstrates how to implement the state machine pattern using FMesh components.
// It shows how to:
// 1. Define valid state transitions
// 2. Process events that trigger state changes
// 3. Handle invalid state transitions
// 4. Monitor state changes
// The pattern is useful for:
// - Order processing systems
// - Game state management
// - Task scheduling
// - Workflow engines
func main() {
	// Create components
	generator := component.New("generator").
		WithOutputs("event").
		WithActivationFunc(eventGenerator)

	stateRouter := component.New("state-router").
		WithInputs("event", "current-state").
		WithOutputs("event", "state").
		WithActivationFunc(stateRouter)

	idleState := component.New("idle-state").
		WithInputs("event").
		WithOutputs("state").
		WithActivationFunc(func(this *component.Component) error {
			event := this.InputByName("event").FirstSignalPayloadOrNil()
			if event != nil {
				fmt.Printf("Idle State: Processing event '%s'\n", event)
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		})

	runningState := component.New("running-state").
		WithInputs("event").
		WithOutputs("state").
		WithActivationFunc(func(this *component.Component) error {
			event := this.InputByName("event").FirstSignalPayloadOrNil()
			if event != nil {
				fmt.Printf("Running State: Processing event '%s'\n", event)
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		})

	pausedState := component.New("paused-state").
		WithInputs("event").
		WithOutputs("state").
		WithActivationFunc(func(this *component.Component) error {
			event := this.InputByName("event").FirstSignalPayloadOrNil()
			if event != nil {
				fmt.Printf("Paused State: Processing event '%s'\n", event)
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		})

	stoppedState := component.New("stopped-state").
		WithInputs("event").
		WithOutputs("state").
		WithActivationFunc(func(this *component.Component) error {
			event := this.InputByName("event").FirstSignalPayloadOrNil()
			if event != nil {
				fmt.Printf("Stopped State: Processing event '%s'\n", event)
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		})

	collector := component.New("collector").
		WithInputs("state").
		WithActivationFunc(func(this *component.Component) error {
			state := this.InputByName("state").FirstSignalPayloadOrNil()
			if state != nil {
				fmt.Printf("State Machine: Current state is %s\n", state)
			}
			time.Sleep(500 * time.Millisecond)
			return nil
		})

	// Connect components
	generator.OutputByName("event").PipeTo(stateRouter.InputByName("event"))
	stateRouter.OutputByName("event").PipeTo(idleState.InputByName("event"))
	stateRouter.OutputByName("event").PipeTo(runningState.InputByName("event"))
	stateRouter.OutputByName("event").PipeTo(pausedState.InputByName("event"))
	stateRouter.OutputByName("event").PipeTo(stoppedState.InputByName("event"))

	idleState.OutputByName("state").PipeTo(stateRouter.InputByName("current-state"))
	runningState.OutputByName("state").PipeTo(stateRouter.InputByName("current-state"))
	pausedState.OutputByName("state").PipeTo(stateRouter.InputByName("current-state"))
	stoppedState.OutputByName("state").PipeTo(stateRouter.InputByName("current-state"))

	stateRouter.OutputByName("state").PipeTo(collector.InputByName("state"))

	// Initialize state machine in Idle state
	stateRouter.InputByName("current-state").PutSignals(signal.New("Idle"))

	// Create and run the mesh
	mesh := fmesh.New("state-machine-example").
		WithDescription("Demonstrates state machine pattern").
		WithComponents(generator, stateRouter, idleState, runningState, pausedState, stoppedState, collector)

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}

func eventGenerator(this *component.Component) error {
	events := []string{"start", "pause", "resume", "stop", "restart"}

	for _, event := range events {
		fmt.Printf("Generator: Sending event: %s\n", event)
		this.OutputByName("event").PutSignals(signal.New(event))
		time.Sleep(2 * time.Second)
	}

	// Wait a bit longer to ensure all events are processed
	time.Sleep(5 * time.Second)
	return nil
}

func stateRouter(this *component.Component) error {
	event := this.InputByName("event").FirstSignalPayloadOrNil()
	currentState := this.InputByName("current-state").FirstSignalPayloadOrDefault("Idle").(string)

	if event == nil {
		return nil
	}

	fmt.Printf("State Router: Processing event '%s' in state '%s'\n", event, currentState)
	newState := currentState

	// Route event to appropriate state based on current state
	switch currentState {
	case "Idle":
		if event == "start" {
			this.OutputByName("event").PutSignals(signal.New(event))
			newState = "Running"
			fmt.Printf("State Machine: Transitioning from Idle to Running\n")
		}
	case "Running":
		if event == "pause" {
			this.OutputByName("event").PutSignals(signal.New(event))
			newState = "Paused"
			fmt.Printf("State Machine: Transitioning from Running to Paused\n")
		} else if event == "stop" {
			this.OutputByName("event").PutSignals(signal.New(event))
			newState = "Stopped"
			fmt.Printf("State Machine: Transitioning from Running to Stopped\n")
		}
	case "Paused":
		if event == "resume" {
			this.OutputByName("event").PutSignals(signal.New(event))
			newState = "Running"
			fmt.Printf("State Machine: Transitioning from Paused to Running\n")
		}
	case "Stopped":
		if event == "restart" {
			this.OutputByName("event").PutSignals(signal.New(event))
			newState = "Idle"
			fmt.Printf("State Machine: Transitioning from Stopped to Idle\n")
		}
	}

	this.OutputByName("state").PutSignals(signal.New(newState))
	time.Sleep(500 * time.Millisecond)
	return nil
}
