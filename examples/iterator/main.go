package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Iterator Pattern Example
// This example demonstrates how to implement the iterator pattern using FMesh components.
// It shows how to:
// 1. Define different iteration strategies
// 2. Access collection elements without exposing internal structure
// 3. Support multiple traversal methods
// 4. Encapsulate iteration logic
// The pattern is useful for:
// - Accessing collection elements sequentially
// - Supporting different traversal methods
// - Hiding collection implementation details
// - Providing a uniform interface for iteration
// Common use cases include:
// - Tree traversal
// - Graph traversal
// - Collection iteration
// - File system navigation
// - Database result sets
func main() {
	// Create data generator
	generator := component.New("generator").
		WithDescription("Generates collection of items").
		WithInputs("start").
		WithOutputs("collection").
		WithActivationFunc(func(this *component.Component) error {
			items := []string{
				"Item1", "Item2", "Item3", "Item4", "Item5",
			}
			this.OutputByName("collection").PutSignals(signal.New(items))
			return nil
		})

	// Create forward iterator
	forwardIterator := component.New("forward-iterator").
		WithDescription("Iterates through items in forward order").
		WithInputs("collection").
		WithOutputs("item").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("collection").AllSignalsOrNil() {
				items := s.PayloadOrNil().([]string)
				fmt.Printf("Forward Iterator: Starting forward traversal\n")

				for i := 0; i < len(items); i++ {
					fmt.Printf("Forward Iterator: Processing item %d: %s\n", i+1, items[i])
					this.OutputByName("item").PutSignals(signal.New(fmt.Sprintf("Forward: %s", items[i])))
					time.Sleep(100 * time.Millisecond)
				}
			}
			return nil
		})

	// Create reverse iterator
	reverseIterator := component.New("reverse-iterator").
		WithDescription("Iterates through items in reverse order").
		WithInputs("collection").
		WithOutputs("item").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("collection").AllSignalsOrNil() {
				items := s.PayloadOrNil().([]string)
				fmt.Printf("Reverse Iterator: Starting reverse traversal\n")

				for i := len(items) - 1; i >= 0; i-- {
					fmt.Printf("Reverse Iterator: Processing item %d: %s\n", i+1, items[i])
					this.OutputByName("item").PutSignals(signal.New(fmt.Sprintf("Reverse: %s", items[i])))
					time.Sleep(100 * time.Millisecond)
				}
			}
			return nil
		})

	// Create alternating iterator
	alternatingIterator := component.New("alternating-iterator").
		WithDescription("Iterates through items in alternating order").
		WithInputs("collection").
		WithOutputs("item").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("collection").AllSignalsOrNil() {
				items := s.PayloadOrNil().([]string)
				fmt.Printf("Alternating Iterator: Starting alternating traversal\n")

				left := 0
				right := len(items) - 1
				count := 1

				for left <= right {
					if left == right {
						fmt.Printf("Alternating Iterator: Processing middle item: %s\n", items[left])
						this.OutputByName("item").PutSignals(signal.New(fmt.Sprintf("Alternating: %s", items[left])))
					} else {
						fmt.Printf("Alternating Iterator: Processing items %d and %d: %s, %s\n",
							count, count+1, items[left], items[right])
						this.OutputByName("item").PutSignals(signal.New(fmt.Sprintf("Alternating: %s, %s",
							items[left], items[right])))
						count += 2
					}
					left++
					right--
					time.Sleep(100 * time.Millisecond)
				}
			}
			return nil
		})

	// Create result collector
	collector := component.New("collector").
		WithDescription("Collects iteration results").
		WithInputs("item").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("item").AllSignalsOrNil() {
				fmt.Printf("Collector: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("iterator-example").
		WithDescription("Demonstrates iterator pattern").
		WithComponents(generator, forwardIterator, reverseIterator, alternatingIterator, collector)

	// Connect components
	generator.OutputByName("collection").PipeTo(forwardIterator.InputByName("collection"))
	generator.OutputByName("collection").PipeTo(reverseIterator.InputByName("collection"))
	generator.OutputByName("collection").PipeTo(alternatingIterator.InputByName("collection"))

	forwardIterator.OutputByName("item").PipeTo(collector.InputByName("item"))
	reverseIterator.OutputByName("item").PipeTo(collector.InputByName("item"))
	alternatingIterator.OutputByName("item").PipeTo(collector.InputByName("item"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
