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

Inspired by [J. Paul Morrison's FBP](https://jpaulm.github.io/fbp/), F-Mesh brings dataflow programming to Go with a small, deliberately untyped API — one pipe can carry anything, so you model the flow instead of the type graph.

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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Create the components
	concat, err := component.New("concat",
		component.WithInputs("i1", "i2"),
		component.WithOutputs("res"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			word1 := signal.AsOrDefault(this.InputByName("i1").Signals().First(), "")
			word2 := signal.AsOrDefault(this.InputByName("i2").Signals().First(), "")
			return this.OutputByName("res").PutSignals(signal.New(word1 + word2))
		}))
	if err != nil {
		return err
	}

	uppercase, err := component.New("uppercase",
		component.WithInputs("i1"),
		component.WithOutputs("res"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			str := signal.AsOrDefault(this.InputByName("i1").Signals().First(), "")
			return this.OutputByName("res").PutSignals(signal.New(strings.ToUpper(str)))
		}))
	if err != nil {
		return err
	}

	// Create the mesh
	fm, err := fmesh.New("hello world")
	if err != nil {
		return err
	}
	if err = fm.AddComponents(concat, uppercase); err != nil {
		return err
	}

	// Connect components via a pipe
	if err = concat.OutputByName("res").PipeTo(uppercase.InputByName("i1")); err != nil {
		return err
	}

	// Set initial inputs
	if err = concat.InputByName("i1").PutSignals(signal.New("hello ")); err != nil {
		return err
	}
	if err = concat.InputByName("i2").PutSignals(signal.New("world!")); err != nil {
		return err
	}

	// Run the mesh. Cancelling ctx stops it at the next cycle boundary.
	if _, err = fm.Run(ctx); err != nil {
		return err
	}

	// Get the result
	result, err := uppercase.OutputByName("res").Signals().FirstPayload()
	if err != nil {
		return err
	}
	fmt.Printf("Result: %v\n", result) // Result: HELLO WORLD!
	return nil
}
```

---

## Key Features

### **Component-Based Architecture**
Build complex workflows from simple, reusable components. Each component is independent and testable.

### **Hooks System**
Extend behavior at any execution point - mesh lifecycle, cycles, component activations, and port operations:

```go
fm.SetupHooks(func(h *fmesh.Hooks) {
    h.BeforeRun(func(ctx context.Context, fm *fmesh.FMesh) error {
        fmt.Println("Starting mesh...")
        return nil
    })
    h.AfterCycle(func(ctx context.Context, hookCtx *fmesh.CycleContext) error {
        fmt.Printf("Cycle #%d complete\n", hookCtx.Cycle.Number())
        return nil
    })
})
```

### **Runtime Observability**
`Run(ctx)` returns a `RuntimeInfo` report with per-cycle activation results and timing — history retention is configurable for long runs.

### **Metadata & Filtering**
Tag signals, components, and ports with labels (string) and scalars (numeric), then filter, route, and aggregate them with consistent collection APIs.

### **Discrete Time Model**
Components activate in cycles (artificial "time"), allowing multiple components to process simultaneously - like lighting multiple lamps at once.

### **Deterministic Runs**
Same input, same output, every time — given activation functions that are themselves deterministic. Signals keep their arrival order within a port, upstream components are drained in name order, and a component's ports are traversed in name order. Reproducible runs make meshes testable, which dataflow systems usually are not.

### **Untyped by Design**
Signals carry `any`. One pipe can hold a string, a struct and an error at once, which is the thing Go channels cannot do and the reason to use a mesh at all. The cost is honest: type mismatches surface when you read a payload, not when you compile.

```go
n, err := signal.As[int](sig)          // reports a mismatch — prefer this
n := signal.AsOrDefault(sig, 0)        // swallows it; use only when a fallback is genuinely right
```

### **Concurrency Out of the Box**
All components in a single activation cycle run concurrently - no need to manage goroutines or other concurrency primitives yourself.

### **Cancellation & Deadlines**
`Run(ctx)` takes a context and passes it to every activation function and hook. Cancel it to stop the mesh at the next cycle boundary; a configured `TimeLimit` becomes a deadline on that context, so it reaches the HTTP calls and queries inside your components instead of only being checked between cycles.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    ...
})

_, err := fm.Run(ctx) // errors.Is(err, fmesh.ErrRunCanceled) once cancel() lands
```

Cancellation is cooperative: Go cannot preempt a goroutine, so an activation function that ignores its context still runs to completion.

---

## Core Concepts

| Concept | Description |
|---------|-------------|
| **[Component](https://github.com/hovsep/fmesh/wiki/301.-Component)** | The main building block - has inputs, outputs, and an activation function |
| **[Port](https://github.com/hovsep/fmesh/wiki/302.-Ports)** | Entry/exit points on components. Unlimited inputs and outputs per component |
| **[Pipe](https://github.com/hovsep/fmesh/wiki/303.-Pipes)** | Connects an output port to an input port to transfer data |
| **[Signal](https://github.com/hovsep/fmesh/wiki/201.-Signals)** | Data packets flowing through pipes. Type-agnostic with optional labels |
| **[Cycle](https://github.com/hovsep/fmesh/wiki/401.-Scheduling-rules)** | One "tick" of execution where all ready components activate |

---

## Four rules worth knowing up front

These are not footnotes. Each one will surprise you exactly once, and the wiki covers them in full.

**1. A component activates only when an input port has signals.** There are no source components; a
mesh is started by seeding inputs. Forgetting to seed produces a run that does nothing and returns
`nil` — that is the rule working, not a failure.

```go
_ = producer.InputByName("start").PutSignals(signal.New("go")) // without this, nothing happens
```

**2. Order is by name.** Signals keep arrival order within a port; upstream components drain in
component-name order; a component's own ports are traversed in port-name order. If a component
reads all its inputs at once, **the port names decide the order** — so name them for the order you
want.

**3. Payloads are shared, so treat them as read-only.** Fan-out hands the same `*Signal` to every
destination, and those destinations activate concurrently. Read what arrives and produce new
signals; never mutate a received map, slice or pointer. Run your mesh tests with `-race`.

**4. Cancellation is cooperative.** `Run(ctx)` stops the mesh at the next cycle boundary, and
`TimeLimit` becomes a deadline on that context. Go cannot preempt a goroutine, so an activation
function that ignores its context still blocks the mesh until it returns.

Full detail: [401. Scheduling rules](https://github.com/hovsep/fmesh/wiki/401.-Scheduling-rules)
and [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats).

---

## Use Cases

F-Mesh excels at:

- **Data transformation pipelines** - ETL, data processing, format conversion
- **Workflow automation** - Multi-step business processes
- **Computational graphs** - Scientific computing, simulations
- **Game logic** - Entity systems, behavior trees
- **Batch event processing** - Bounded sets of events, processed to completion
- **Experimental architectures** - Prototyping dataflow designs

It is **not** a streaming engine: a mesh runs to completion over the data it was given, rather than staying up and consuming an endless stream. See [Limitations](#limitations).

---

## Documentation

- **[Wiki](https://github.com/hovsep/fmesh/wiki)** - Full documentation (source lives in [`docs/wiki`](docs/wiki) — edit via PR, it is auto-synced to the wiki)
- **[Examples Repository](https://github.com/hovsep/fmesh-examples)** - Working examples and patterns
- **[API Reference](https://pkg.go.dev/github.com/hovsep/fmesh)** - Complete API docs
- **[Flow-Based Programming](https://jpaulm.github.io/fbp/)** - Learn about FBP (by J. Paul Morrison)
---

## Limitations

F-Mesh is **not** a classical FBP implementation:

- Not suitable for long-running components
- No wall-clock time events (timers, tickers)
- Components execute in discrete cycles, not real-time
- No backpressure: a mesh holds its signals in memory for the whole run

For real-time streaming or long-running processes, consider alternatives like traditional FBP systems or message queues.

Known trade-offs that can bite — partial output before a wait, shared payloads, work lost when one
component errors — are written down in [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats)
rather than left for you to find.

---

## Versioning

F-Mesh is pre-production and the API is still moving. **Minor versions may contain breaking
changes** until this notice is removed — pin an exact version if that matters to you, and read
[CHANGELOG.md](CHANGELOG.md), where breaking changes are listed explicitly per release.

---

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md), then:

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
