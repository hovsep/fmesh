# autowire — pipes by naming convention

A mesh of any size accumulates long stretches of wiring that say nothing a reader could not have
guessed: everything that keeps time gets the clock, everything that wants the weather gets the
weather. Written out, that is a list to maintain, and forgetting an entry fails silently — the
component sits there, looking perfectly connected, waiting on inputs that never arrive.

`autowire.Plugin` makes the list derivable. Give it a rule mapping an output port to the input
port name that should receive it, and it wires every match as components arrive:

```go
import "github.com/hovsep/fmesh/plugin/autowire"

fm, err := fmesh.New("mesh",
    fmesh.WithPlugins(autowire.Broadcast("time")),
)

// No PipeTo calls: every component with an input named "time"
// is wired to every output named "time".
err = fm.AddComponents(clock, heart, lung)
```

## Conventions

Three ready-made rules plus the general form:

| Constructor | Rule |
|-------------|------|
| `autowire.Broadcast("time")` | every output named `time` → every input named `time` |
| `autowire.BroadcastAs("tick", "time")` | every output named `tick` → every input named `time` |
| `autowire.Prefixed("env_")` | an output → the input named `<prefix><component>_<port>`, e.g. `env_sun_uvi` |
| `&autowire.Plugin{InputNameFor: func(source *component.Component, output *port.Port) string {...}}` | your own rule; return `""` to decline |

Wiring happens on every arrival, **in both directions**: a new component is offered every output
already in the mesh, and every output it brings is offered to the components already there. So the
order of `AddComponents` does not matter — which is the property that makes a convention safe to
lean on.

## Several conventions at once

A mesh often wants more than one rule — a clock reaching everything that keeps time, *and* a set
of environmental factors reaching whatever asked for them by name. Those are separate rules, so
they are separate plugins:

```go
fm, err := fmesh.New("habitat", fmesh.WithPlugins(
    autowire.BroadcastAs("tick", "time"),
    autowire.Prefixed("env_"),
))
```

Each carries its own `PluginName`, so they coexist. Two conventions that happen to agree on the
same output/input pair wire it once, not twice.

## Limits

- A component is never wired to **itself**. A convention is not how you express a loopback — use
  `c.LoopbackPipe(out, in)`.
- Ports must exist **before** `AddComponents`: arrival is the only moment a component is looked
  at, so ports added later with `AddInputs`/`AddOutputs` are never wired, and nothing reports it.

## More

- [Plugins wiki page](https://github.com/hovsep/fmesh/wiki/502.-Plugins) — how mesh plugins work
  and when to write your own.
- [API reference](https://pkg.go.dev/github.com/hovsep/fmesh/plugin/autowire).
