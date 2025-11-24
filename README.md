<div align="center">
  <img src="./assets/img/logo.png" width="200" height="200" alt="f-mesh"/>
  <h1>F-Mesh</h1>
  <p><em>Flow-Based Programming framework for Go</em></p>
[![Go Report Card](https://goreportcard.com/badge/github.com/hovsep/fmesh)](https://goreportcard.com/report/github.com/hovsep/fmesh)
[![Go Reference](https://pkg.go.dev/badge/github.com/hovsep/fmesh.svg)](https://pkg.go.dev/github.com/hovsep/fmesh)
[![Latest Release](https://img.shields.io/github/v/release/hovsep/fmesh)](https://github.com/hovsep/fmesh/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/hovsep/fmesh/branch/main/graph/badge.svg)](https://codecov.io/gh/hovsep/fmesh)
</div>
---

## What is F-Mesh?

F-Mesh is a **Flow-Based Programming (FBP)** framework that lets you build applications as a graph of independent, reusable components. Think of it as connecting building blocks with pipes - data flows through your program like water through a network of connected components.

Inspired by [J. Paul Morrison's FBP](https://jpaulm.github.io/fbp/), F-Mesh brings the power of dataflow programming to Go with a clean, type-safe API.

<img src="https://github.com/user-attachments/assets/045bb7ac-0852-4a0d-9158-6af2d6e66dbb" width="500px">

---

## Installation

```bash
go get github.com/hovsep/fmesh
```

---

## Quick Start

Here's a simple mesh that concatenates two strings and converts them to uppercase:

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	// Create the mesh
	fm := fmesh.New("hello world").
		AddComponents(
			component.New("concat").
				AddInputs("i1", "i2").
				AddOutputs("res").
				WithActivationFunc(func(c *component.Component) error {
					word1 := c.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
					word2 := c.InputByName("i2").Signals().FirstPayloadOrDefault("").(string)
					c.OutputByName("res").PutSignals(signal.New(word1 + word2))
					return nil
				}),
			component.New("uppercase").
				AddInputs("i1").
				AddOutputs("res").
				WithActivationFunc(func(c *component.Component) error {
					str := c.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
					c.OutputByName("res").PutSignals(signal.New(strings.ToUpper(str)))
					return nil
				}),
		)

	// Connect components via pipes
	fm.ComponentByName("concat").OutputByName("res").
		PipeTo(fm.ComponentByName("uppercase").InputByName("i1"))

	// Set initial inputs
	fm.ComponentByName("concat").InputByName("i1").PutSignals(signal.New("hello "))
	fm.ComponentByName("concat").InputByName("i2").PutSignals(signal.New("world!"))

	// Run the mesh
	_, err := fm.Run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Get results
	results := fm.ComponentByName("uppercase").OutputByName("res").Signals().FirstPayloadOrNil()
	fmt.Printf("Result: %v\n", results) // Output: HELLO WORLD!
}
```

---

## Key Features

### **Component-Based Architecture**
Build complex workflows from simple, reusable components. Each component is independent and testable.

### **Hooks System** _(New in v1.4.0)_
Extend behavior at any execution point - mesh lifecycle, cycles, component activations, and port operations:

```go
fm.SetupHooks(func(h *fmesh.Hooks) {
    h.BeforeRun(func(fm *fmesh.FMesh) error {
        fmt.Println("Starting mesh...")
        return nil
    })
    h.CycleEnd(func(ctx *fmesh.CycleContext) error {
        fmt.Printf("Cycle #%d complete\n", ctx.Cycle.Number())
        return nil
    })
})
```

### **Labels & Filtering**
Tag signals and components with labels, then filter and process them with powerful collection APIs.

### **Discrete Time Model**
Components activate in cycles (artificial "time"), allowing multiple components to process simultaneously - like lighting multiple lamps at once.

### **Chainable API**
Fluent interface for building meshes with readable, declarative code.

### **Concurrency Out of the Box**
All components in a single activation cycle run concurrently - no need to manage goroutines or other concurrency primitives yourself.

---

## Core Concepts

| Concept | Description |
|---------|-------------|
| **[Component](https://github.com/hovsep/fmesh/wiki/5.-Component)** | The main building block - has inputs, outputs, and an activation function |
| **[Port](https://github.com/hovsep/fmesh/wiki/3.-Ports)** | Entry/exit points on components. Unlimited inputs and outputs per component |
| **[Pipe](https://github.com/hovsep/fmesh/wiki/4.-Pipes)** | Connects an output port to an input port to transfer data |
| **[Signal](https://github.com/hovsep/fmesh/wiki/2.-Signals)** | Data packets flowing through pipes. Type-agnostic with optional labels |
| **[Cycle](https://github.com/hovsep/fmesh/wiki/6.-Scheduling-rules)** | One "tick" of execution where all ready components activate |

---

## Use Cases

F-Mesh excels at:

- **Data transformation pipelines** - ETL, data processing, format conversion
- **Workflow automation** - Multi-step business processes
- **Computational graphs** - Scientific computing, simulations
- **Game logic** - Entity systems, behavior trees
- **Stream processing** - Event handling, reactive systems
- **Experimental architectures** - Prototyping dataflow designs

---

## Documentation

- **[Wiki](https://github.com/hovsep/fmesh/wiki)** - Full documentation
- **[Examples Repository](https://github.com/hovsep/fmesh-examples)** - Working examples and patterns
- **[API Reference](https://pkg.go.dev/github.com/hovsep/fmesh)** - Complete API docs
- **[Flow-Based Programming](https://jpaulm.github.io/fbp/)** - Learn about FBP (by J. Paul Morrison)
---

## Limitations

F-Mesh is **not** a classical FBP implementation:

- Not suitable for long-running components
- No wall-clock time events (timers, tickers)
- Components execute in discrete cycles, not real-time

For real-time streaming or long-running processes, consider alternatives like traditional FBP systems or message queues.

---

## Contributing

Contributions are welcome! Please:

1. Check existing [issues](https://github.com/hovsep/fmesh/issues) or create a new one
2. Fork the repository
3. Create a feature branch
4. Submit a pull request

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">
  <p>Made by <a href="https://github.com/hovsep">@hovsep</a></p>
  <p>Star us on GitHub if you find F-Mesh useful!</p>
</div>
