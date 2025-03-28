package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Mediator Pattern Example
// This example demonstrates how to implement the mediator pattern using FMesh components.
// It shows how to:
// 1. Centralize communication between components
// 2. Reduce direct dependencies between components
// 3. Implement complex communication protocols
// 4. Handle broadcast and targeted messages
// The pattern is useful for:
// - Decoupling components in a system
// - Managing complex interactions
// - Implementing communication protocols
// - Coordinating multiple components
// Common use cases include:
// - Chat applications
// - Air traffic control systems
// - Event handling systems
// - GUI component interactions
// - Multi-player game coordination
func main() {
	// Create user components that will communicate through the mediator
	user1 := component.New("user1").
		WithDescription("First user in the chat system").
		WithInputs("start", "message", "broadcast").
		WithOutputs("send").
		WithActivationFunc(func(this *component.Component) error {
			// Start sending messages when triggered
			for range this.InputByName("start").AllSignalsOrNil() {
				// Send messages
				messages := []string{
					"TO:user2:Hello from User1",
					"BROADCAST:User1 to all: Hi everyone",
					"TO:user3:Direct message to User3",
				}

				for _, msg := range messages {
					this.OutputByName("send").PutSignals(signal.New(msg))
					time.Sleep(100 * time.Millisecond)
				}
			}

			// Handle received messages
			for _, s := range this.InputByName("message").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User1 received direct message: %s\n", msg)
			}

			// Handle broadcast messages
			for _, s := range this.InputByName("broadcast").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User1 received broadcast: %s\n", msg)
			}

			return nil
		})

	user2 := component.New("user2").
		WithDescription("Second user in the chat system").
		WithInputs("start", "message", "broadcast").
		WithOutputs("send").
		WithActivationFunc(func(this *component.Component) error {
			// Start sending messages when triggered
			for range this.InputByName("start").AllSignalsOrNil() {
				// Send messages
				messages := []string{
					"TO:user1:Hello back from User2",
					"BROADCAST:User2 to all: Hello everyone",
				}

				for _, msg := range messages {
					this.OutputByName("send").PutSignals(signal.New(msg))
					time.Sleep(100 * time.Millisecond)
				}
			}

			// Handle received messages
			for _, s := range this.InputByName("message").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User2 received direct message: %s\n", msg)
			}

			// Handle broadcast messages
			for _, s := range this.InputByName("broadcast").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User2 received broadcast: %s\n", msg)
			}

			return nil
		})

	user3 := component.New("user3").
		WithDescription("Third user in the chat system").
		WithInputs("start", "message", "broadcast").
		WithOutputs("send").
		WithActivationFunc(func(this *component.Component) error {
			// Start sending messages when triggered
			for range this.InputByName("start").AllSignalsOrNil() {
				// Send messages
				messages := []string{
					"TO:user1:Response from User3",
					"BROADCAST:User3 to all: Greetings",
				}

				for _, msg := range messages {
					this.OutputByName("send").PutSignals(signal.New(msg))
					time.Sleep(100 * time.Millisecond)
				}
			}

			// Handle received messages
			for _, s := range this.InputByName("message").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User3 received direct message: %s\n", msg)
			}

			// Handle broadcast messages
			for _, s := range this.InputByName("broadcast").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				fmt.Printf("User3 received broadcast: %s\n", msg)
			}

			return nil
		})

	// Create mediator component
	mediator := component.New("mediator").
		WithDescription("Central mediator for communication").
		WithInputs("user1", "user2", "user3").
		WithOutputs("user1_msg", "user1_broadcast", "user2_msg", "user2_broadcast", "user3_msg", "user3_broadcast").
		WithActivationFunc(func(this *component.Component) error {
			// Process messages from User1
			for _, s := range this.InputByName("user1").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				if msg[:9] == "BROADCAST" {
					broadcast := msg[10:]
					this.OutputByName("user2_broadcast").PutSignals(signal.New(broadcast))
					this.OutputByName("user3_broadcast").PutSignals(signal.New(broadcast))
				} else if msg[:2] == "TO" {
					parts := msg[3:]
					target := parts[:5]
					content := parts[6:]
					if target == "user2" {
						this.OutputByName("user2_msg").PutSignals(signal.New(content))
					} else if target == "user3" {
						this.OutputByName("user3_msg").PutSignals(signal.New(content))
					}
				}
			}

			// Process messages from User2
			for _, s := range this.InputByName("user2").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				if msg[:9] == "BROADCAST" {
					broadcast := msg[10:]
					this.OutputByName("user1_broadcast").PutSignals(signal.New(broadcast))
					this.OutputByName("user3_broadcast").PutSignals(signal.New(broadcast))
				} else if msg[:2] == "TO" {
					parts := msg[3:]
					target := parts[:5]
					content := parts[6:]
					if target == "user1" {
						this.OutputByName("user1_msg").PutSignals(signal.New(content))
					} else if target == "user3" {
						this.OutputByName("user3_msg").PutSignals(signal.New(content))
					}
				}
			}

			// Process messages from User3
			for _, s := range this.InputByName("user3").AllSignalsOrNil() {
				msg := s.PayloadOrNil().(string)
				if msg[:9] == "BROADCAST" {
					broadcast := msg[10:]
					this.OutputByName("user1_broadcast").PutSignals(signal.New(broadcast))
					this.OutputByName("user2_broadcast").PutSignals(signal.New(broadcast))
				} else if msg[:2] == "TO" {
					parts := msg[3:]
					target := parts[:5]
					content := parts[6:]
					if target == "user1" {
						this.OutputByName("user1_msg").PutSignals(signal.New(content))
					} else if target == "user2" {
						this.OutputByName("user2_msg").PutSignals(signal.New(content))
					}
				}
			}
			return nil
		})

	// Create starter component
	starter := component.New("starter").
		WithDescription("Triggers users to start communication").
		WithInputs("start").
		WithOutputs("trigger").
		WithActivationFunc(func(this *component.Component) error {
			for range this.InputByName("start").AllSignalsOrNil() {
				this.OutputByName("trigger").PutSignals(signal.New("start"))
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("mediator-example").
		WithDescription("Demonstrates mediator pattern").
		WithComponents(starter, user1, user2, user3, mediator)

	// Connect starter to users
	starter.OutputByName("trigger").PipeTo(user1.InputByName("start"))
	starter.OutputByName("trigger").PipeTo(user2.InputByName("start"))
	starter.OutputByName("trigger").PipeTo(user3.InputByName("start"))

	// Connect components through the mediator
	user1.OutputByName("send").PipeTo(mediator.InputByName("user1"))
	user2.OutputByName("send").PipeTo(mediator.InputByName("user2"))
	user3.OutputByName("send").PipeTo(mediator.InputByName("user3"))

	mediator.OutputByName("user1_msg").PipeTo(user1.InputByName("message"))
	mediator.OutputByName("user1_broadcast").PipeTo(user1.InputByName("broadcast"))
	mediator.OutputByName("user2_msg").PipeTo(user2.InputByName("message"))
	mediator.OutputByName("user2_broadcast").PipeTo(user2.InputByName("broadcast"))
	mediator.OutputByName("user3_msg").PipeTo(user3.InputByName("message"))
	mediator.OutputByName("user3_broadcast").PipeTo(user3.InputByName("broadcast"))

	// Start the mesh by triggering the starter
	starter.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
