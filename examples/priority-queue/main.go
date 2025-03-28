package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

type PriorityMessage struct {
	Priority int
	Content  string
}

// Priority Queue Pattern Example
// This example demonstrates how to implement a priority queue using FMesh components.
// It shows how to:
// 1. Generate messages with different priority levels
// 2. Sort messages by priority (highest first)
// 3. Process messages in priority order
// The pattern is useful when you need to handle tasks with different urgency levels,
// ensuring that high-priority tasks are processed before low-priority ones.
func main() {
	// Create message generator
	generator := component.New("generator").
		WithDescription("Generates messages with different priorities").
		WithInputs("start").
		WithOutputs("messages").
		WithActivationFunc(func(this *component.Component) error {
			messages := []PriorityMessage{
				{Priority: 1, Content: "Low priority task"},
				{Priority: 3, Content: "High priority task"},
				{Priority: 2, Content: "Medium priority task"},
				{Priority: 3, Content: "Urgent task"},
				{Priority: 1, Content: "Regular task"},
			}

			for _, msg := range messages {
				this.Logger().Printf("Generating: %s (Priority: %d)", msg.Content, msg.Priority)
				this.OutputByName("messages").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create priority queue component
	queue := component.New("priority-queue").
		WithDescription("Processes messages based on priority").
		WithInputs("messages").
		WithOutputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			var messages []PriorityMessage
			for _, sig := range this.InputByName("messages").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(PriorityMessage)
				messages = append(messages, msg)
			}

			// Sort by priority (highest first)
			for i := 0; i < len(messages)-1; i++ {
				for j := i + 1; j < len(messages); j++ {
					if messages[i].Priority < messages[j].Priority {
						messages[i], messages[j] = messages[j], messages[i]
					}
				}
			}

			// Process messages in priority order
			for _, msg := range messages {
				this.Logger().Printf("Processing: %s (Priority: %d)", msg.Content, msg.Priority)
				this.OutputByName("processed").PutSignals(signal.New(msg))
				time.Sleep(200 * time.Millisecond)
			}
			return nil
		})

	// Create processor component
	processor := component.New("processor").
		WithDescription("Processes messages in priority order").
		WithInputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("processed").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(PriorityMessage)
				this.Logger().Printf("Completed: %s (Priority: %d)", msg.Content, msg.Priority)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("priority-queue-example").
		WithDescription("Demonstrates priority queue pattern using FMesh").
		WithComponents(generator, queue, processor)

	// Connect components
	generator.OutputByName("messages").PipeTo(queue.InputByName("messages"))
	queue.OutputByName("processed").PipeTo(processor.InputByName("processed"))

	// Start the mesh with initial signal
	generator.InputByName("start").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
