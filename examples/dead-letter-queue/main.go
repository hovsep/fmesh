package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

type Message struct {
	Content string
	Retries int
}

// Dead Letter Queue Pattern Example
// This example demonstrates how to implement a dead letter queue (DLQ) pattern using FMesh components.
// It shows how to:
// 1. Process messages with potential failures
// 2. Implement retry logic for failed messages
// 3. Move messages that exceed retry limits to a dead letter queue
// The pattern is useful for handling message processing failures in a graceful way,
// ensuring that problematic messages don't block the main processing flow and can be
// analyzed later for debugging or manual intervention.
func main() {
	// Create message generator
	generator := component.New("generator").
		WithDescription("Generates messages to be processed").
		WithInputs("start").
		WithOutputs("messages").
		WithActivationFunc(func(this *component.Component) error {
			messages := []Message{
				{Content: "Valid message", Retries: 0},
				{Content: "Invalid message", Retries: 0},
				{Content: "Another valid message", Retries: 0},
				{Content: "Error message", Retries: 0},
			}

			for _, msg := range messages {
				this.Logger().Printf("Generating: %s", msg.Content)
				this.OutputByName("messages").PutSignals(signal.New(msg))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create processor component
	processor := component.New("processor").
		WithDescription("Processes messages and handles failures").
		WithInputs("messages").
		WithOutputs("processed", "failed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("messages").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(Message)
				if strings.Contains(msg.Content, "Invalid") || strings.Contains(msg.Content, "Error") {
					this.Logger().Printf("Processing failed: %s", msg.Content)
					this.OutputByName("failed").PutSignals(signal.New(msg))
				} else {
					this.Logger().Printf("Processing successful: %s", msg.Content)
					this.OutputByName("processed").PutSignals(signal.New(msg))
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create retry handler
	retryHandler := component.New("retry-handler").
		WithDescription("Handles failed messages with retries").
		WithInputs("failed").
		WithOutputs("retry", "dlq").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("failed").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(Message)
				msg.Retries++
				if msg.Retries <= 2 {
					this.Logger().Printf("Retrying message: %s (attempt %d)", msg.Content, msg.Retries)
					this.OutputByName("retry").PutSignals(signal.New(msg))
				} else {
					this.Logger().Printf("Message moved to DLQ: %s", msg.Content)
					this.OutputByName("dlq").PutSignals(signal.New(msg))
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create DLQ handler
	dlqHandler := component.New("dlq-handler").
		WithDescription("Handles messages that exceeded retry limit").
		WithInputs("dlq").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("dlq").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(Message)
				this.Logger().Printf("DLQ: %s (failed after %d retries)", msg.Content, msg.Retries)
			}
			return nil
		})

	// Create success handler
	successHandler := component.New("success-handler").
		WithDescription("Handles successfully processed messages").
		WithInputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("processed").AllSignalsOrNil() {
				msg := sig.PayloadOrNil().(Message)
				this.Logger().Printf("Successfully processed: %s", msg.Content)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("dead-letter-queue-example").
		WithDescription("Demonstrates dead letter queue pattern using FMesh").
		WithComponents(generator, processor, retryHandler, dlqHandler, successHandler)

	// Connect components
	generator.OutputByName("messages").PipeTo(processor.InputByName("messages"))
	processor.OutputByName("processed").PipeTo(successHandler.InputByName("processed"))
	processor.OutputByName("failed").PipeTo(retryHandler.InputByName("failed"))
	retryHandler.OutputByName("retry").PipeTo(processor.InputByName("messages"))
	retryHandler.OutputByName("dlq").PipeTo(dlqHandler.InputByName("dlq"))

	// Start the mesh with initial signal
	generator.InputByName("start").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
