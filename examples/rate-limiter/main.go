package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

type RateLimitState struct {
	LastRequest  time.Time
	RequestCount int
}

func main() {
	// Create request generator
	generator := component.New("generator").
		WithDescription("Generates requests to be rate limited").
		WithOutputs("requests").
		WithActivationFunc(func(this *component.Component) error {
			for i := 0; i < 10; i++ {
				req := fmt.Sprintf("Request %d", i+1)
				this.Logger().Printf("Generating request: %s", req)
				this.OutputByName("requests").PutSignals(signal.New(req))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create rate limiter component
	rateLimiter := component.New("rate-limiter").
		WithDescription("Implements rate limiting").
		WithInputs("requests").
		WithOutputs("allowed", "rejected").
		WithActivationFunc(func(this *component.Component) error {
			state := RateLimitState{
				LastRequest:  time.Now(),
				RequestCount: 0,
			}

			// Rate limit: 3 requests per second
			rateLimit := 3
			window := time.Second

			for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
				req := sig.PayloadOrNil().(string)
				now := time.Now()

				// Reset counter if window has passed
				if now.Sub(state.LastRequest) > window {
					state.RequestCount = 0
					state.LastRequest = now
				}

				if state.RequestCount < rateLimit {
					state.RequestCount++
					this.Logger().Printf("Request allowed: %s (count: %d)", req, state.RequestCount)
					this.OutputByName("allowed").PutSignals(signal.New(req))
				} else {
					this.Logger().Printf("Request rejected: %s (rate limit exceeded)", req)
					this.OutputByName("rejected").PutSignals(signal.New(req))
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create allowed request handler
	allowedHandler := component.New("allowed-handler").
		WithDescription("Handles allowed requests").
		WithInputs("requests").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
				req := sig.PayloadOrNil().(string)
				this.Logger().Printf("Processing allowed request: %s", req)
				time.Sleep(200 * time.Millisecond)
			}
			return nil
		})

	// Create rejected request handler
	rejectedHandler := component.New("rejected-handler").
		WithDescription("Handles rejected requests").
		WithInputs("requests").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName("requests").AllSignalsOrNil() {
				req := sig.PayloadOrNil().(string)
				this.Logger().Printf("Request rejected due to rate limit: %s", req)
			}
			return nil
		})

	// Create the mesh
	mesh := fmesh.New("rate-limiter-example").
		WithDescription("Demonstrates rate limiter pattern using FMesh").
		WithComponents(generator, rateLimiter, allowedHandler, rejectedHandler)

	// Connect components
	generator.OutputByName("requests").PipeTo(rateLimiter.InputByName("requests"))
	rateLimiter.OutputByName("allowed").PipeTo(allowedHandler.InputByName("requests"))
	rateLimiter.OutputByName("rejected").PipeTo(rejectedHandler.InputByName("requests"))

	// Start the mesh with initial signal
	generator.InputByName("requests").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
