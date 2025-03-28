package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Pub-Sub (Publisher-Subscriber) Pattern Example
// This example demonstrates how to implement a publish-subscribe pattern using FMesh components.
// It shows how to:
// 1. Create a publisher that sends messages to different topics
// 2. Create subscribers that listen to specific topics
// 3. Route messages to appropriate subscribers based on topics
// The pattern is useful for decoupling message producers from consumers and
// implementing topic-based message routing.
// Common use cases include:
// - News/content distribution systems
// - Event notification systems
// - Real-time data feeds
// - Message brokers
func main() {
	// Create publisher component
	publisher := component.New("publisher").
		WithDescription("Publishes messages to topics").
		WithInputs("start").
		WithOutputs("news", "sports", "tech").
		WithActivationFunc(func(this *component.Component) error {
			messages := []struct {
				topic   string
				content string
			}{
				{"news", "Breaking news: Major announcement"},
				{"sports", "Match results: Team A vs Team B"},
				{"tech", "New gadget released"},
				{"news", "Weather update"},
				{"sports", "Tournament schedule"},
				{"tech", "Software update available"},
			}

			for _, msg := range messages {
				this.Logger().Printf("Publishing to %s: %s", msg.topic, msg.content)
				this.OutputByName(msg.topic).PutSignals(signal.New(msg.content))
				time.Sleep(200 * time.Millisecond)
			}
			return nil
		})

	// Create subscriber components
	subscribers := make([]*component.Component, 3)
	topics := []string{"news", "sports", "tech"}

	for i, topic := range topics {
		subscriberID := fmt.Sprintf("subscriber-%s", topic)
		subscribers[i] = component.New(subscriberID).
			WithDescription(fmt.Sprintf("Subscribes to %s topic", topic)).
			WithInputs(topic).
			WithActivationFunc(func(this *component.Component) error {
				for _, sig := range this.InputByName(topic).AllSignalsOrNil() {
					msg := sig.PayloadOrNil().(string)
					this.Logger().Printf("%s received: %s", this.Name(), msg)
					time.Sleep(100 * time.Millisecond)
				}
				return nil
			})
	}

	// Create the mesh
	mesh := fmesh.New("pub-sub-example").
		WithDescription("Demonstrates pub-sub pattern using FMesh").
		WithComponents(append([]*component.Component{publisher}, subscribers...)...)

	// Connect components
	for i, subscriber := range subscribers {
		publisher.OutputByName(topics[i]).PipeTo(subscriber.InputByName(topics[i]))
	}

	// Start the mesh with initial signal
	publisher.InputByName("start").PutSignals(signal.New("start"))
	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
