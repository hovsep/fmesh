package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

type CircuitState struct {
	State     string
	Failures  int
	LastError string
}

// Circuit Breaker Pattern Example
// This example demonstrates how to implement the circuit breaker pattern using FMesh components.
// It shows how to:
// 1. Monitor failures in a service
// 2. Open the circuit after a threshold of failures is reached
// 3. Reject requests when the circuit is open
// 4. Handle successful and failed requests differently
// The pattern is useful for preventing cascading failures in distributed systems by
// failing fast and providing fallback options.
// Common use cases include:
// - Protecting external service calls
// - Database connection management
// - Rate limiting and backoff strategies
// - System resilience and fault tolerance
func main() {
	// Create request generator
	generator := component.New("generator").
		WithDescription("Generates requests to test circuit breaker").
		WithInputs("start").
		WithOutputs("request").
		WithActivationFunc(func(this *component.Component) error {
			for i := 0; i < 10; i++ {
				this.OutputByName("request").PutSignals(signal.New(fmt.Sprintf("Request %d", i)))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Initialize circuit breaker state
	state := &CircuitState{
		State:     "CLOSED",
		Failures:  0,
		LastError: "",
	}

	// Create circuit breaker
	circuitBreaker := component.New("circuit-breaker").
		WithDescription("Implements circuit breaker pattern").
		WithInputs("request").
		WithOutputs("response", "state").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("request").AllSignalsOrNil() {
				// If circuit is open, reject the request
				if state.State == "OPEN" {
					this.OutputByName("response").PutSignals(signal.New(fmt.Sprintf("Circuit OPEN - Request %s rejected", s.PayloadOrNil())))
					continue
				}

				// Simulate service call with random failures
				if time.Now().UnixNano()%2 == 0 {
					state.Failures++
					state.LastError = "Service error"

					// If failures exceed threshold, open the circuit
					if state.Failures >= 3 {
						state.State = "OPEN"
						this.OutputByName("state").PutSignals(signal.New(fmt.Sprintf("Circuit state changed to OPEN after %d failures", state.Failures)))
					}

					this.OutputByName("response").PutSignals(signal.New(fmt.Sprintf("Request %s failed: %s", s.PayloadOrNil(), state.LastError)))
				} else {
					// Successful request
					state.Failures = 0
					state.LastError = ""
					this.OutputByName("response").PutSignals(signal.New(fmt.Sprintf("Request %s processed successfully", s.PayloadOrNil())))
				}
			}
			return nil
		})

	// Create response monitor
	monitor := component.New("monitor").
		WithDescription("Monitors circuit breaker responses and state changes").
		WithInputs("response", "state").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("response").AllSignalsOrNil() {
				fmt.Printf("Monitor received response: %v\n", s.PayloadOrNil())
			}
			for _, s := range this.InputByName("state").AllSignalsOrNil() {
				fmt.Printf("Monitor received state change: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("circuit-breaker-example").
		WithDescription("Demonstrates circuit breaker pattern").
		WithComponents(generator, circuitBreaker, monitor)

	// Connect components
	generator.OutputByName("request").PipeTo(circuitBreaker.InputByName("request"))
	circuitBreaker.OutputByName("response").PipeTo(monitor.InputByName("response"))
	circuitBreaker.OutputByName("state").PipeTo(monitor.InputByName("state"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
