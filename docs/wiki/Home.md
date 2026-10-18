> This wiki is synced from [`docs/wiki`](https://github.com/hovsep/fmesh/tree/main/docs/wiki) in the main repository. Edits made here are overwritten on the next sync. Edit by PR instead.

F-Mesh (also FMesh or fmesh) is a Go framework for Flow-Based Programming (FBP). You build a
data-flow network from components. Each component has input and output **ports**. Type-agnostic
**pipes** connect the ports, and **signals** flow through them.

Extension points: hooks (run your code at key moments), plugins, metadata (string, bool and number tags
on every entity) and collections/groups (work with many entities at once).

# Installation

```
go get github.com/hovsep/fmesh
```

Requires Go 1.27 or later.

# Versioning

F-Mesh uses semantic versioning. While it is pre-production, **minor versions may contain breaking
changes**. [CHANGELOG.md](https://github.com/hovsep/fmesh/blob/main/CHANGELOG.md) lists each one.
Pin an exact version if that matters to you.

Release tags are plain semver (`v1.12.0`). Each release is named after an Armenian place; the name
is in the release title, not the tag. See the
[releases page](https://github.com/hovsep/fmesh/releases).

# User guide

| Topic | Page |
|-------|------|
| Quick start | [101. Quick start](https://github.com/hovsep/fmesh/wiki/101.-Quick-start) |
| Signals | [201. Signals](https://github.com/hovsep/fmesh/wiki/201.-Signals) |
| Metadata | [202. Metadata](https://github.com/hovsep/fmesh/wiki/202.-Metadata) |
| Collections & groups | [203. Collections and Groups](https://github.com/hovsep/fmesh/wiki/203.-Collections-and-Groups) |
| Components | [301. Component](https://github.com/hovsep/fmesh/wiki/301.-Component) |
| Ports | [302. Ports](https://github.com/hovsep/fmesh/wiki/302.-Ports) |
| Pipes | [303. Pipes](https://github.com/hovsep/fmesh/wiki/303.-Pipes) |
| Running the mesh | [401. Scheduling rules](https://github.com/hovsep/fmesh/wiki/401.-Scheduling-rules) |
| Inspecting a run | [402. Inspecting a run](https://github.com/hovsep/fmesh/wiki/402.-Inspecting-a-run) |
| Hooks | [501. Hooks](https://github.com/hovsep/fmesh/wiki/501.-Hooks) |
| Plugins | [502. Plugins](https://github.com/hovsep/fmesh/wiki/502.-Plugins) |
| Configuration & tips | [601. Tips & tricks](https://github.com/hovsep/fmesh/wiki/601.-Tips-&-tricks) |
| Patterns & recipes | [602. Patterns & recipes](https://github.com/hovsep/fmesh/wiki/602.-Patterns-and-recipes) |
| Caveats & trade-offs | [603. Caveats](https://github.com/hovsep/fmesh/wiki/603.-Caveats) |
| Export / visualization | [701. Export](https://github.com/hovsep/fmesh/wiki/701.-Export) |

# API reference (user-facing packages)

* [fmesh](https://pkg.go.dev/github.com/hovsep/fmesh): mesh, options, hooks
* [component](https://pkg.go.dev/github.com/hovsep/fmesh/component)
* [port](https://pkg.go.dev/github.com/hovsep/fmesh/port)
* [signal](https://pkg.go.dev/github.com/hovsep/fmesh/signal)
* [meta](https://pkg.go.dev/github.com/hovsep/fmesh/meta)
* [profiler](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/profiler): bundled mesh plugin
* [autowire](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/autowire): bundled mesh plugin
* [fmesh-export](https://github.com/hovsep/fmesh-export): JSON, YAML, DOT, Mermaid, D2 and PlantUML exporters, and the `export.Exporter` interface

The `cycle` package appears in the run report: see [402. Inspecting a run](https://github.com/hovsep/fmesh/wiki/402.-Inspecting-a-run).

# Examples

[fmesh-examples](https://github.com/hovsep/fmesh-examples) has runnable programs, grouped by what they teach. Start at the top.

**Basics**

- [string_processing](https://github.com/hovsep/fmesh-examples/tree/main/basics/string_processing): two components and one pipe, the smallest useful mesh
- [filter](https://github.com/hovsep/fmesh-examples/tree/main/basics/filter): metadata-based content routing
- [pipeline](https://github.com/hovsep/fmesh-examples/tree/main/basics/pipeline): a multi-stage chain that reads stdin and files

**Patterns: how a mesh is wired and driven**

- [fibonacci](https://github.com/hovsep/fmesh-examples/tree/main/patterns/fibonacci): a generator built on a loopback pipe
- [nesting](https://github.com/hovsep/fmesh-examples/tree/main/patterns/nesting): a mesh running inside a component
- [load_balancer](https://github.com/hovsep/fmesh-examples/tree/main/patterns/load_balancer): round-robin dispatch/collect with indexed ports and component state
- [async_input](https://github.com/hovsep/fmesh-examples/tree/main/patterns/async_input): a mesh driven from outside, signals injected on a ticker and results drained into a channel
- [retry](https://github.com/hovsep/fmesh-examples/tree/main/patterns/retry): `WithRetry` and `WithRetryIf`, temporary errors retried with backoff, permanent ones not
- [state_machine](https://github.com/hovsep/fmesh-examples/tree/main/patterns/state_machine): states as components, transitions as pipes, the current state in mesh metadata
- [autodiff](https://github.com/hovsep/fmesh-examples/tree/main/patterns/autodiff): a neural network learning XOR, values flowing forward and gradients back
- [tree_search](https://github.com/hovsep/fmesh-examples/tree/main/patterns/tree_search): one component per tree node, a search going down one level per cycle and merging on the way up
- [hypergraph](https://github.com/hovsep/fmesh-examples/tree/main/patterns/hypergraph): many-to-many edges, a reaction firing only when every reactant is there
- [fpga](https://github.com/hovsep/fmesh-examples/tree/main/patterns/fpga): generic logic cells flashed at runtime from a Verilog-like file
- [unix_pipes](https://github.com/hovsep/fmesh-examples/tree/main/patterns/unix_pipes): unix commands as components, one log analysed three ways at once
- [rule_engine](https://github.com/hovsep/fmesh-examples/tree/main/patterns/rule_engine): rules from a config file compiled into a mesh on every run

**Simulations: systems evolving over time**

- [electric_circuit](https://github.com/hovsep/fmesh-examples/tree/main/simulation/electric_circuit): a stateful feedback loop that stops on its own
- [can_bus](https://github.com/hovsep/fmesh-examples/tree/main/simulation/can_bus): a broadcast bus; the advanced variant models the full CAN stack
- [combustion_engine](https://github.com/hovsep/fmesh-examples/tree/main/simulation/combustion_engine): a four-cylinder engine driven from the browser
- [elevator](https://github.com/hovsep/fmesh-examples/tree/main/simulation/elevator): two elevators down to the last button and brake, every click a signal
- [restaurant](https://github.com/hovsep/fmesh-examples/tree/main/simulation/restaurant): one day in a restaurant, one mesh run per minute
- [life](https://github.com/hovsep/fmesh-examples/tree/main/simulation/life): a large body simulation with nested meshes and a live TUI

**Graphics**

- [graphviz](https://github.com/hovsep/fmesh-examples/tree/main/graphics/graphviz): static and per-cycle graph export
- [ray_tracer](https://github.com/hovsep/fmesh-examples/tree/main/graphics/ray_tracer): a wavefront 3D ray tracer with parallel tile bands and a reflection cycle

The techniques they use are listed in [602. Patterns & recipes](https://github.com/hovsep/fmesh/wiki/602.-Patterns-and-recipes).
