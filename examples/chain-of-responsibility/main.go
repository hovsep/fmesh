package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Chain of Responsibility Pattern Example
// This example demonstrates how to implement the chain of responsibility pattern using FMesh components.
// It shows how to:
// 1. Create a chain of handlers that process requests in sequence
// 2. Pass requests through the chain until they are handled
// 3. Allow handlers to decide whether to process a request or pass it to the next handler
// 4. Monitor the flow of requests through the chain
// The pattern is useful for building flexible processing pipelines where:
// - Multiple handlers may process a request
// - The handler that can process a request isn't known in advance
// - The chain of handlers can be modified dynamically
// Common use cases include:
// - Request filtering and validation
// - Authentication and authorization
// - Data transformation pipelines
// - Event processing systems
// - Logging and monitoring chains
func main() {
	// Create request generator component
	generator := component.New("generator").
		WithDescription("Generates sample requests").
		WithInputs("start").
		WithOutputs("request").
		WithActivationFunc(func(this *component.Component) error {
			requests := []string{
				"valid_user:valid_data",     // Valid request
				"invalid_user:valid_data",   // Invalid auth
				"valid_user:invalid_data",   // Invalid data
				"admin_user:sensitive_data", // Special processing
			}

			for _, req := range requests {
				this.OutputByName("request").PutSignals(signal.New(req))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create authentication handler
	authHandler := component.New("auth-handler").
		WithDescription("Authenticates requests").
		WithInputs("request").
		WithOutputs("authenticated", "rejected").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("request").AllSignalsOrNil() {
				req := s.PayloadOrNil().(string)
				fmt.Printf("Auth handler processing: %s\n", req)

				parts := strings.Split(req, ":")
				user := parts[0]

				if user == "valid_user" || user == "admin_user" {
					this.OutputByName("authenticated").PutSignals(signal.New(req))
					fmt.Printf("Request authenticated: %s\n", req)
				} else {
					this.OutputByName("rejected").PutSignals(signal.New(fmt.Sprintf("Authentication failed for: %s", req)))
					fmt.Printf("Request rejected by auth: %s\n", req)
				}
			}
			return nil
		})

	// Create validation handler
	validationHandler := component.New("validation-handler").
		WithDescription("Validates request data").
		WithInputs("request").
		WithOutputs("validated", "rejected").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("request").AllSignalsOrNil() {
				req := s.PayloadOrNil().(string)
				fmt.Printf("Validation handler processing: %s\n", req)

				parts := strings.Split(req, ":")
				data := parts[1]

				if data == "valid_data" || data == "sensitive_data" {
					this.OutputByName("validated").PutSignals(signal.New(req))
					fmt.Printf("Request validated: %s\n", req)
				} else {
					this.OutputByName("rejected").PutSignals(signal.New(fmt.Sprintf("Validation failed for: %s", req)))
					fmt.Printf("Request rejected by validation: %s\n", req)
				}
			}
			return nil
		})

	// Create processing handler
	processingHandler := component.New("processing-handler").
		WithDescription("Processes validated requests").
		WithInputs("request").
		WithOutputs("processed", "special").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("request").AllSignalsOrNil() {
				req := s.PayloadOrNil().(string)
				fmt.Printf("Processing handler received: %s\n", req)

				parts := strings.Split(req, ":")
				user := parts[0]

				if user == "admin_user" {
					this.OutputByName("special").PutSignals(signal.New(fmt.Sprintf("Special processing for: %s", req)))
					fmt.Printf("Request sent for special processing: %s\n", req)
				} else {
					this.OutputByName("processed").PutSignals(signal.New(fmt.Sprintf("Successfully processed: %s", req)))
					fmt.Printf("Request processed normally: %s\n", req)
				}
			}
			return nil
		})

	// Create monitor for rejected requests
	rejectionMonitor := component.New("rejection-monitor").
		WithDescription("Monitors rejected requests").
		WithInputs("rejected").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("rejected").AllSignalsOrNil() {
				fmt.Printf("Rejection monitor: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create monitor for processed requests
	processMonitor := component.New("process-monitor").
		WithDescription("Monitors processed requests").
		WithInputs("processed", "special").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("processed").AllSignalsOrNil() {
				fmt.Printf("Process monitor (normal): %v\n", s.PayloadOrNil())
			}
			for _, s := range this.InputByName("special").AllSignalsOrNil() {
				fmt.Printf("Process monitor (special): %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("chain-of-responsibility-example").
		WithDescription("Demonstrates chain of responsibility pattern").
		WithComponents(generator, authHandler, validationHandler, processingHandler, rejectionMonitor, processMonitor)

	// Connect components to form the chain
	generator.OutputByName("request").PipeTo(authHandler.InputByName("request"))
	authHandler.OutputByName("authenticated").PipeTo(validationHandler.InputByName("request"))
	validationHandler.OutputByName("validated").PipeTo(processingHandler.InputByName("request"))

	// Connect rejection handlers
	authHandler.OutputByName("rejected").PipeTo(rejectionMonitor.InputByName("rejected"))
	validationHandler.OutputByName("rejected").PipeTo(rejectionMonitor.InputByName("rejected"))

	// Connect process monitors
	processingHandler.OutputByName("processed").PipeTo(processMonitor.InputByName("processed"))
	processingHandler.OutputByName("special").PipeTo(processMonitor.InputByName("special"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
