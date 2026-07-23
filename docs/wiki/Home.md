> 📝 This wiki is auto-synced from [`docs/wiki`](https://github.com/hovsep/fmesh/tree/main/docs/wiki) in the main repository. Do not edit pages here — changes will be overwritten on the next sync. Edit via PR instead.

F-Mesh (aka FMesh or fmesh) is a Golang framework inspired by FBP (Flow-Based Programming) that enables the creation of data flow networks using interconnected components, each processing signals. Components may have multiple inputs and outputs called "ports" linked via type-agnostic pipes.

F-Mesh provides extension mechanisms including hooks (observability), labels and scalars (metadata), and collections/groups (working with multiple entities).

# Installation

```
go get github.com/hovsep/fmesh
```

# Release naming and versioning conventions

F-Mesh releases are named after 17 historical capitals of Armenia, honouring the ancient cities that played foundational roles in Armenian history. This tradition highlights the project's growth with each version, paralleling Armenia's own historical progression.

F-Mesh follows semantic versioning. See [releases page](https://github.com/hovsep/fmesh/releases) for the latest version and changelog.

# User guide

| Topic | Page |
|-------|------|
| Quick start | [101. Quick start](https://github.com/hovsep/fmesh/wiki/101.-Quick-start) |
| Signals | [201. Signals](https://github.com/hovsep/fmesh/wiki/201.-Signals) |
| Labels & scalars | [202. Metadata](https://github.com/hovsep/fmesh/wiki/202.-Metadata) |
| Collections & groups | [203. Collections and Groups](https://github.com/hovsep/fmesh/wiki/203.-Collections-and-Groups) |
| Components | [301. Component](https://github.com/hovsep/fmesh/wiki/301.-Component) |
| Ports | [302. Ports](https://github.com/hovsep/fmesh/wiki/302.-Ports) |
| Pipes | [303. Pipes](https://github.com/hovsep/fmesh/wiki/303.-Pipes) |
| Running the mesh | [401. Scheduling rules](https://github.com/hovsep/fmesh/wiki/401.-Scheduling-rules) |
| Inspecting a run | [402. Inspecting a run](https://github.com/hovsep/fmesh/wiki/402.-Inspecting-a-run) |
| Hooks | [501. Hooks](https://github.com/hovsep/fmesh/wiki/501.-Hooks) |
| Configuration & tips | [601. Tips & tricks](https://github.com/hovsep/fmesh/wiki/601.-Tips-&-tricks) |
| Export / visualization | [701. Export](https://github.com/hovsep/fmesh/wiki/701.-Export) |

# API reference (user-facing packages)

* [FMesh](https://pkg.go.dev/github.com/hovsep/fmesh) — mesh, config, hooks
* [Component](https://pkg.go.dev/github.com/hovsep/fmesh/component)
* [Port](https://pkg.go.dev/github.com/hovsep/fmesh/port)
* [Signal](https://pkg.go.dev/github.com/hovsep/fmesh/signal)
* [Meta (labels & scalars)](https://pkg.go.dev/github.com/hovsep/fmesh/meta)

The `cycle` package surfaces in the run-time report — see [402. Inspecting a run](https://github.com/hovsep/fmesh/wiki/402.-Inspecting-a-run); hooks are covered in [501. Hooks](https://github.com/hovsep/fmesh/wiki/501.-Hooks).

# Examples

[fmesh-examples](https://github.com/hovsep/fmesh-examples) — including [nested meshes](https://github.com/hovsep/fmesh-examples/tree/main/nesting), pipelines, filters, and more.
