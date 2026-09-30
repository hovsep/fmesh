# autowire: pipes by naming convention

Big meshes have a lot of obvious wiring: everything that keeps time gets the clock, everything that
needs the weather gets the weather. Written by hand, that is a list to maintain. Forget one entry
and nothing fails: the component just waits for inputs that never come.

`autowire.Plugin` derives that list from a rule. The rule maps an output port to the name of the
input port that should receive it. The plugin wires every match as components are added:

```go
import "github.com/hovsep/fmesh/plugin/autowire"

fm, err := fmesh.New("mesh",
    fmesh.WithPlugins(autowire.Broadcast("time")),
)

// No PipeTo calls: every output named "time" is piped
// to every input named "time".
err = fm.AddComponents(clock, heart, lung)
```

## Conventions

| Constructor | Rule |
|-------------|------|
| `autowire.Broadcast("time")` | every output named `time` → every input named `time` |
| `autowire.BroadcastAs("tick", "time")` | every output named `tick` → every input named `time` |
| `autowire.Prefixed("env_")` | an output → the input named `<prefix><component>_<port>`, e.g. `env_sun_uvi` |
| `&autowire.Plugin{InputNameFor: func(source *component.Component, output *port.Port) string {...}}` | your own rule; return `""` to skip |

Wiring runs **in both directions** each time a component is added. The new component's inputs
are matched against outputs already in the mesh, and its outputs against inputs already there. So
the order of `AddComponents` does not matter.

## Several conventions at once

One plugin holds one rule. For more rules, register more plugins:

```go
fm, err := fmesh.New("habitat", fmesh.WithPlugins(
    autowire.BroadcastAs("tick", "time"),
    autowire.Prefixed("env_"),
))
```

Each has its own `PluginName`, so they coexist. When two rules match the same output/input pair,
the pipe is created once, not twice.

## Limits

- A component is never wired to **itself**. For a loopback, use `c.LoopbackPipe(out, in)`.
- Ports must exist **before** `AddComponents`. The plugin looks at a component only when it is
  added, so ports added later with `AddInputs`/`AddOutputs` are never wired, and nothing reports it.

## More

- [Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins): how mesh plugins work
  and when to write your own.
- [API reference](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/autowire).
