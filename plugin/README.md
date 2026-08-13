# Bundled plugins

`plugin` holds no code of its own — each bundled mesh plugin is its own package, with its own
README:

| Plugin | What it does |
|--------|--------------|
| [`autowire`](autowire/README.md) | Pipes components together by naming convention instead of by hand — wiring derived on arrival, in both directions, so `AddComponents` order does not matter. |
| [`profiler`](profiler/README.md) | Mesh-centric measurement: run/cycle/activation timing, per-pipe throughput (never-fired pipes included), and a per-cycle timeline. Every number is attributable to the mesh. |

Both attach the same way:

```go
fm, err := fmesh.New("mesh", fmesh.WithPlugins(
    profiler.New(),
    autowire.Broadcast("time"),
))
```

To write your own — the interface is two methods, `Name()` and `Init(*fmesh.FMesh) error` — see
the [Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins), which covers the
`OnComponentAdded` pattern every mesh plugin builds on.
