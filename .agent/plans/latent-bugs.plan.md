# Latent Bugs

Bugs and API hazards found during the chainable-error refactor audit.
Independent of that refactor — can land in any order.

---

## Bug 1 — Global mutable `*Config` pointer

**File:** `config.go:33`

```go
var defaultConfig = &Config{ ... }
```

`New()` assigns `fm.config = defaultConfig` (same pointer).
Any code path that mutates a field through that pointer (e.g. `fm.config.Logger = ...`)
modifies the shared default, affecting all subsequent `New()` calls in the same process.

**Fix:** Replace the package-level var with a constructor function:

```go
func newDefaultConfig() *Config {
    return &Config{
        ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
        CyclesLimit:           1000,
        Debug:                 false,
        Logger:                getDefaultLogger(),
        TimeLimit:             UnlimitedTime,
    }
}
```

`New()` calls `newDefaultConfig()` to get a fresh copy per instance.

---

## Bug 2 — `LoopbackPipe` silently discards errors / nil-pointer panic

**File:** `component/io.go:211`

```go
func (c *Component) LoopbackPipe(out, in string) {
    c.OutputByName(out).PipeTo(c.InputByName(in))
}
```

If `out` or `in` does not name an existing port, `OutputByName`/`InputByName` returns `nil`.
Calling `.PipeTo(...)` on a nil `*Port` panics. The error from `PipeTo` is also silently
discarded (no return value).

**Fix:** Return `error` and guard nil lookups:

```go
func (c *Component) LoopbackPipe(out, in string) error {
    outPort := c.OutputByName(out)
    if outPort == nil {
        return fmt.Errorf("loopback pipe: output port %q not found", out)
    }
    inPort := c.InputByName(in)
    if inPort == nil {
        return fmt.Errorf("loopback pipe: input port %q not found", in)
    }
    return outPort.PipeTo(inPort)
}
```

Note: Phase 3 of `remove-chainable-errors.plan.md` also changes `LoopbackPipe` to return
`error`. If that plan lands first, apply only the nil-guard additions here.

---

## Bug 3 — Typed-nil trap on `ParentComponent()` interface assertion

**File:** `fmesh.go:442`

```go
destComponent := destPort.ParentComponent().(*component.Component)
```

`ParentComponent()` returns a `port.ParentComponent` interface.
The nil check just before (`if destPort.ParentComponent() == nil`) only catches the
*untyped* nil interface. A `*component.Component` typed nil stored in the interface passes
the check and then panics at the type assertion on the next line.

**Fix:** Assert first, then nil-check the concrete value:

```go
parent, ok := destPort.ParentComponent().(*component.Component)
if !ok || parent == nil {
    continue
}
destComponent := parent
```

---

## Checklist

- [ ] Bug 1: `defaultConfig` → `newDefaultConfig()` in `config.go`
- [ ] Bug 2: `LoopbackPipe` nil guard + `error` return in `component/io.go`
- [ ] Bug 3: Typed-nil guard for `ParentComponent()` assertion in `fmesh.go:442`
- [ ] `make test && make lint && make fmt` green after all three fixes
