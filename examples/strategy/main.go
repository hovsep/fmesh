package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Strategy Pattern Example
// This example demonstrates how to implement the strategy pattern using FMesh components.
// It shows how to:
// 1. Define different algorithms (strategies) for a specific task
// 2. Make algorithms interchangeable at runtime
// 3. Select appropriate strategy based on context
// 4. Encapsulate algorithm-specific logic
// The pattern is useful for:
// - Supporting multiple algorithms for a task
// - Allowing runtime algorithm selection
// - Isolating algorithm variations
// - Enabling easy addition of new algorithms
// Common use cases include:
// - Payment processing
// - Data compression
// - Sorting algorithms
// - Authentication methods
// - Route calculation
func main() {
	// Create order generator
	generator := component.New("generator").
		WithDescription("Generates orders with different payment methods").
		WithInputs("start").
		WithOutputs("order").
		WithActivationFunc(func(this *component.Component) error {
			orders := []string{
				"ORDER1:Credit Card:500.00",
				"ORDER2:PayPal:150.50",
				"ORDER3:Crypto:1000.00",
				"ORDER4:Bank Transfer:750.25",
			}

			for _, order := range orders {
				this.OutputByName("order").PutSignals(signal.New(order))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create credit card payment processor
	creditCardProcessor := component.New("credit-card-processor").
		WithDescription("Processes credit card payments").
		WithInputs("payment").
		WithOutputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("payment").AllSignalsOrNil() {
				payment := s.PayloadOrNil().(string)
				fmt.Printf("Credit Card Processor: Processing payment for %s\n", payment)
				fmt.Printf("- Validating card details\n")
				fmt.Printf("- Checking available credit\n")
				fmt.Printf("- Processing transaction\n")
				fmt.Printf("- Sending confirmation to card issuer\n")

				this.OutputByName("result").PutSignals(signal.New(fmt.Sprintf("Credit Card payment completed for %s", payment)))
			}
			return nil
		})

	// Create PayPal payment processor
	paypalProcessor := component.New("paypal-processor").
		WithDescription("Processes PayPal payments").
		WithInputs("payment").
		WithOutputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("payment").AllSignalsOrNil() {
				payment := s.PayloadOrNil().(string)
				fmt.Printf("PayPal Processor: Processing payment for %s\n", payment)
				fmt.Printf("- Connecting to PayPal API\n")
				fmt.Printf("- Verifying PayPal account\n")
				fmt.Printf("- Initiating transfer\n")
				fmt.Printf("- Waiting for confirmation\n")

				this.OutputByName("result").PutSignals(signal.New(fmt.Sprintf("PayPal payment completed for %s", payment)))
			}
			return nil
		})

	// Create crypto payment processor
	cryptoProcessor := component.New("crypto-processor").
		WithDescription("Processes cryptocurrency payments").
		WithInputs("payment").
		WithOutputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("payment").AllSignalsOrNil() {
				payment := s.PayloadOrNil().(string)
				fmt.Printf("Crypto Processor: Processing payment for %s\n", payment)
				fmt.Printf("- Generating wallet address\n")
				fmt.Printf("- Waiting for blockchain confirmation\n")
				fmt.Printf("- Verifying transaction hash\n")
				fmt.Printf("- Converting to fiat value\n")

				this.OutputByName("result").PutSignals(signal.New(fmt.Sprintf("Crypto payment completed for %s", payment)))
			}
			return nil
		})

	// Create bank transfer processor
	bankProcessor := component.New("bank-processor").
		WithDescription("Processes bank transfer payments").
		WithInputs("payment").
		WithOutputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("payment").AllSignalsOrNil() {
				payment := s.PayloadOrNil().(string)
				fmt.Printf("Bank Transfer Processor: Processing payment for %s\n", payment)
				fmt.Printf("- Validating bank details\n")
				fmt.Printf("- Initiating wire transfer\n")
				fmt.Printf("- Checking transfer status\n")
				fmt.Printf("- Recording transaction details\n")

				this.OutputByName("result").PutSignals(signal.New(fmt.Sprintf("Bank transfer completed for %s", payment)))
			}
			return nil
		})

	// Create payment router
	router := component.New("payment-router").
		WithDescription("Routes payments to appropriate processor").
		WithInputs("order").
		WithOutputs("credit_card", "paypal", "crypto", "bank").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("order").AllSignalsOrNil() {
				order := s.PayloadOrNil().(string)
				parts := strings.Split(order, ":")
				method := parts[1]

				switch method {
				case "Credit Card":
					this.OutputByName("credit_card").PutSignals(signal.New(order))
				case "PayPal":
					this.OutputByName("paypal").PutSignals(signal.New(order))
				case "Crypto":
					this.OutputByName("crypto").PutSignals(signal.New(order))
				case "Bank Transfer":
					this.OutputByName("bank").PutSignals(signal.New(order))
				}
			}
			return nil
		})

	// Create result collector
	collector := component.New("collector").
		WithDescription("Collects payment processing results").
		WithInputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("result").AllSignalsOrNil() {
				fmt.Printf("Payment Result: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("strategy-example").
		WithDescription("Demonstrates strategy pattern").
		WithComponents(generator, router, creditCardProcessor, paypalProcessor, cryptoProcessor, bankProcessor, collector)

	// Connect components
	generator.OutputByName("order").PipeTo(router.InputByName("order"))

	router.OutputByName("credit_card").PipeTo(creditCardProcessor.InputByName("payment"))
	router.OutputByName("paypal").PipeTo(paypalProcessor.InputByName("payment"))
	router.OutputByName("crypto").PipeTo(cryptoProcessor.InputByName("payment"))
	router.OutputByName("bank").PipeTo(bankProcessor.InputByName("payment"))

	creditCardProcessor.OutputByName("result").PipeTo(collector.InputByName("result"))
	paypalProcessor.OutputByName("result").PipeTo(collector.InputByName("result"))
	cryptoProcessor.OutputByName("result").PipeTo(collector.InputByName("result"))
	bankProcessor.OutputByName("result").PipeTo(collector.InputByName("result"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
