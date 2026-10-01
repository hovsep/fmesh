> This wiki is synced from [`docs/wiki`](https://github.com/hovsep/fmesh/tree/main/docs/wiki) in the main repository. Edits made here are overwritten on the next sync. Edit by PR instead.

F-Mesh (also FMesh or fmesh) is a Go framework for Flow-Based Programming (FBP). You build a
data-flow network from components. Each component has input and output **ports**. Type-agnostic
**pipes** connect the ports, and **signals** flow through them.

Extension points: hooks (run your code at key moments), plugins, metadata (string and number tags
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

Release tags are plain semver (`v1.12.0`). Each release is named after one of the 17 historical
capitals of Armenia; the name is in the release title, not the tag. See the
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
* [jsonexport](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/jsonexport): bundled mesh plugin
* [fmesh-export](https://github.com/hovsep/fmesh-export): DOT, Mermaid, D2 and PlantUML exporter plugins

The `cycle` package appears in the run report: see [402. Inspecting a run](https://github.com/hovsep/fmesh/wiki/402.-Inspecting-a-run).

# Examples

[fmesh-examples](https://github.com/hovsep/fmesh-examples) has runnable programs, from small demos to full apps:

- [pipeline](https://github.com/hovsep/fmesh-examples/tree/main/basics/pipeline): a text pipeline built from chained stages
- [filter](https://github.com/hovsep/fmesh-examples/tree/main/basics/filter): metadata-based content routing
- [fibonacci](https://github.com/hovsep/fmesh-examples/tree/main/patterns/fibonacci): a generator built on a loopback pipe
- [electric_circuit](https://github.com/hovsep/fmesh-examples/tree/main/simulation/electric_circuit): a stateful feedback loop that stops on its own
- [load_balancer](https://github.com/hovsep/fmesh-examples/tree/main/patterns/load_balancer): round-robin dispatch/collect with indexed ports
- [async_input](https://github.com/hovsep/fmesh-examples/tree/main/patterns/async_input): a mesh driven by live external input (HTTP crawler)
- [nesting](https://github.com/hovsep/fmesh-examples/tree/main/patterns/nesting): a mesh running inside a component
- [can_bus](https://github.com/hovsep/fmesh-examples/tree/main/simulation/can_bus): a broadcast bus; the advanced variant models the full CAN stack
- [simulation](https://github.com/hovsep/fmesh-examples/tree/main/simulation): a step-by-step simulation driven from a REPL
- [life](https://github.com/hovsep/fmesh-examples/tree/main/simulation/life): a large body simulation with nested meshes and a live TUI
- [graphviz](https://github.com/hovsep/fmesh-examples/tree/main/graphics/graphviz): static and per-cycle graph export

The techniques they use are listed in [602. Patterns & recipes](https://github.com/hovsep/fmesh/wiki/602.-Patterns-and-recipes).
