# Bundled plugins

`plugin` has no code of its own. Each bundled mesh plugin is its own package with its own README:

| Plugin | What it does |
|--------|--------------|
| [`autowire`](autowire/README.md) | Connects components by port-name convention instead of `PipeTo` calls. Wiring works in both directions, so `AddComponents` order does not matter. |
| [`profiler`](profiler/README.md) | Measures the mesh: run, cycle and activation timing, per-pipe throughput (unused pipes included) and a per-cycle timeline. |

Exporters are not plugins: they only read a mesh. They implement the [`export.Exporter`](../export)
interface and live in [fmesh-export](https://github.com/hovsep/fmesh-export): JSON, DOT, Mermaid,
D2 and PlantUML.

All attach the same way:

```go
fm, err := fmesh.New("mesh", fmesh.WithPlugins(
    profiler.New(),
    autowire.Broadcast("time"),
))
```

To write your own, implement two methods: `Name() string` and `Init(*fmesh.FMesh) error`. The
[Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins) shows how, including the
`OnComponentAdded` hook that every mesh plugin builds on.
