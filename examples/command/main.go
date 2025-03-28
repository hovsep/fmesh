package main

import (
	"fmt"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Command Pattern Example
// This example demonstrates how to implement the command pattern using FMesh components.
// It shows how to:
// 1. Encapsulate requests as objects (commands)
// 2. Queue and execute commands
// 3. Support undo/redo operations
// 4. Maintain command history
// The pattern is useful for:
// - Implementing transactional behavior
// - Supporting undo/redo functionality
// - Queueing and scheduling operations
// - Logging and auditing system actions
// Common use cases include:
// - Text editors
// - Drawing applications
// - Database transactions
// - Game action systems
// - Macro recording and playback
func main() {
	// Create command generator
	generator := component.New("generator").
		WithDescription("Generates commands to be executed").
		WithInputs("start").
		WithOutputs("command").
		WithActivationFunc(func(this *component.Component) error {
			commands := []string{
				"ADD:Item1",    // Add an item
				"ADD:Item2",    // Add another item
				"REMOVE:Item1", // Remove first item
				"UNDO",         // Undo last command (REMOVE)
				"ADD:Item3",    // Add a third item
				"REDO",         // Redo the undone command (REMOVE)
			}

			for _, cmd := range commands {
				this.OutputByName("command").PutSignals(signal.New(cmd))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create command executor
	executor := component.New("executor").
		WithDescription("Executes commands and maintains state").
		WithInputs("command").
		WithOutputs("executed", "state").
		WithActivationFunc(func(this *component.Component) error {
			var items []string
			var history []string
			var undoStack []string

			for _, s := range this.InputByName("command").AllSignalsOrNil() {
				cmd := s.PayloadOrNil().(string)
				fmt.Printf("Executing command: %s\n", cmd)

				switch {
				case cmd == "UNDO":
					if len(history) > 0 {
						lastCmd := history[len(history)-1]
						history = history[:len(history)-1]
						undoStack = append(undoStack, lastCmd)

						// Reverse the last command
						if lastCmd[:3] == "ADD" {
							items = items[:len(items)-1]
						} else if lastCmd[:6] == "REMOVE" {
							items = append(items, lastCmd[7:])
						}

						this.OutputByName("executed").PutSignals(signal.New(fmt.Sprintf("Undone: %s", lastCmd)))
					} else {
						this.OutputByName("executed").PutSignals(signal.New("Nothing to undo"))
					}

				case cmd == "REDO":
					if len(undoStack) > 0 {
						redoCmd := undoStack[len(undoStack)-1]
						undoStack = undoStack[:len(undoStack)-1]
						history = append(history, redoCmd)

						// Re-apply the command
						if redoCmd[:3] == "ADD" {
							items = append(items, redoCmd[4:])
						} else if redoCmd[:6] == "REMOVE" {
							for i, item := range items {
								if item == redoCmd[7:] {
									items = append(items[:i], items[i+1:]...)
									break
								}
							}
						}

						this.OutputByName("executed").PutSignals(signal.New(fmt.Sprintf("Redone: %s", redoCmd)))
					} else {
						this.OutputByName("executed").PutSignals(signal.New("Nothing to redo"))
					}

				default:
					// Clear redo stack when new command is executed
					undoStack = nil
					history = append(history, cmd)

					if cmd[:3] == "ADD" {
						items = append(items, cmd[4:])
						this.OutputByName("executed").PutSignals(signal.New(fmt.Sprintf("Added: %s", cmd[4:])))
					} else if cmd[:6] == "REMOVE" {
						for i, item := range items {
							if item == cmd[7:] {
								items = append(items[:i], items[i+1:]...)
								this.OutputByName("executed").PutSignals(signal.New(fmt.Sprintf("Removed: %s", cmd[7:])))
								break
							}
						}
					}
				}

				// Output current state
				this.OutputByName("state").PutSignals(signal.New(fmt.Sprintf("Current items: %v", items)))
			}
			return nil
		})

	// Create monitor
	monitor := component.New("monitor").
		WithDescription("Monitors command execution and state changes").
		WithInputs("executed", "state").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("executed").AllSignalsOrNil() {
				fmt.Printf("Command result: %v\n", s.PayloadOrNil())
			}
			for _, s := range this.InputByName("state").AllSignalsOrNil() {
				fmt.Printf("State update: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("command-example").
		WithDescription("Demonstrates command pattern").
		WithComponents(generator, executor, monitor)

	// Connect components
	generator.OutputByName("command").PipeTo(executor.InputByName("command"))
	executor.OutputByName("executed").PipeTo(monitor.InputByName("executed"))
	executor.OutputByName("state").PipeTo(monitor.InputByName("state"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
