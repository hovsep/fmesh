# Benchmarking & Fuzzing

F-Mesh's priority is *simplicity and a clean API, not performance*. Benchmarks here
are therefore **regression tripwires**, not optimization targets: they exist to catch
a change that accidentally adds an allocation, a copy, or quadratic behavior. Fuzz
targets are **invariant guards** — mainly for the copy-on-write (CoW) rule.

## What we measure & why

Every benchmark reports three numbers (via `b.ReportAllocs()`):

- `ns/op` — wall time. On GitHub-hosted runners (shared VMs) this is **noisy**; trust
  it only as a relative delta over many samples, never as an absolute.
- `B/op` — bytes allocated per op.
- `allocs/op` — heap allocation **count**. This is the most stable metric on noisy CI
  and the best CoW-regression signal: a CoW method's allocation count is effectively an
  assertion about how many copies it makes. A jump here usually means a stray copy — or,
  worse, that someone started mutating in place. Watch it first.

## Comparing runs: benchstat

Never eyeball a single run against another — the variance swamps the signal. Use
[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
go test -run='^$' -bench=. -benchmem -count=8 ./... | tee new.txt
# ...on the other revision...
benchstat old.txt new.txt
```

`-count=8` gives benchstat enough samples to estimate variance. In the output, `~`
means "no statistically significant change" (high p-value) — do not act on it. CI runs
this automatically per PR (`.github/workflows/bench.yml`) and posts the table as a
sticky comment. It is **advisory and never fails the check** — don't gate merges on
runner noise.

## Writing a benchmark

- `b.ReportAllocs()` — always.
- Prefer `for b.Loop() { ... }` (Go 1.24+) over `for range b.N`. `b.Loop()` runs setup
  before the loop outside the timer (no manual `b.ResetTimer()` needed for the common
  case), keeps loop inputs/results alive, and prevents the compiler from eliminating the
  loop body as dead code.
- Keep per-iteration setup out of the measured region. If setup can't be hoisted, guard
  it with `b.StopTimer()`/`b.StartTimer()`.
- Use `b.RunParallel` only when contended concurrency is the thing under test.

## The headline metric: activation cycles per second

The single most important number for f-mesh is **activation cycles per second** — how
fast the scheduler drives full run-loop ticks with every component activating each cycle.
It measures the library overhead added *on top of* the user's activation function, so
`BenchmarkMeshThroughput{Dummy,Bypass}` keep the activation body near-empty and report
custom metrics via `b.ReportMetric`: `cycles/s` and `activations/s` (= cycles/s × size).

Two kinds isolate different overheads, swept over tens/hundreds/thousands of components:

- **dummy** — activation returns `component.ErrWaitingForInputsKeep`, so each component
  keeps its input and re-activates every cycle with no output ports or pipes. This is the
  scheduling/activation floor (goroutine fan-out, `WaitGroup`, result collection).
- **bypass** — activation copies input→output through a self-loop pipe, adding the real
  per-cycle drain/flush/forward cost.

`bypass − dummy` approximates the cost of the signal-movement path. To sustain N cycles
per `Run`, the mesh sets `WithCyclesLimit(N)` + `WithUnlimitedTime()`; the expected stop
error is `ErrReachedMaxAllowedCycles` (not a failure). Inputs are re-primed each `Run`
because the cycle-limit stop skips the final drain and does not carry signal state across
`Run` calls.

## Size sweeps reveal complexity class

A fixed-size benchmark is a single point and hides scaling behavior. Sweep the size so
the shape of the curve is visible:

```go
for _, n := range []int{10, 100, 1_000, 10_000} {
    b.Run(strconv.Itoa(n), func(b *testing.B) { /* ... */ })
}
```

This is how `BenchmarkGroupBuild` exposes that building a group via repeated `With` is
O(n²) (each `With` allocates a fresh slice), and how `BenchmarkMeshRunWide` exercises the
run loop's one-goroutine-per-component-per-cycle cost at scale.

Mesh-scale gotchas:

- **Wide, not deep.** A linear pipeline of N components needs N cycles; N > `CyclesLimit`
  (default 1000) or wall time > `TimeLimit` (default 5s) stops the mesh early. Scale
  benchmarks use *wide* meshes (all components activate in one cycle) or raise the config.
- **Shallow copy is payload-size independent.** `Signal` CoW copies the interface header,
  not the pointed-to payload. `BenchmarkGroupPayloadSize` asserts this by keeping `ns/op`
  flat as payload grows — a tripwire for an accidental deep copy.
- **Fan-in is quadratic** — see "Scaling characteristics" in `runtime.md` for the mechanism
  and practical limits. `BenchmarkMeshFanIn` sweeps this curve; it flattens to linear only
  if forwarding ever batches appends.
- **Long benchmarks accumulate runtime history.** `RuntimeInfo.Cycles` retains every cycle's
  activation results for the whole `Run` (~100 B × components × cycles), so `B/op` in
  sustained-cycle benchmarks includes history growth, not just the signal path.

## When to add one

Add or extend a benchmark when a change touches a hot path — the run loop
(`fmesh.go` `runCycle`/`drainComponents`), the port drain/flush path, or any CoW method
on `signal.Signal`/`signal.Group` — especially if it could add an allocation or a copy.

## Fuzzing

Native Go fuzzing (`func FuzzXxx(f *testing.F)`) guards **properties**, not just crashes.
The high-value property here is the CoW invariant: a mutating-style method must return a
new value and leave its receiver untouched (payload, labels, scalars). Targets live
beside the code they guard (`signal/signal_fuzz_test.go`, `signal/group_fuzz_test.go`).

- Seed corpus runs under normal `go test ./...` (fast, deterministic) — treat it like any
  table test.
- Explore for real with a time budget:
  `go test -run='^$' -fuzz=FuzzSignalCoW -fuzztime=30s ./signal/`
- Discovered failing inputs are written to `testdata/fuzz/<Target>/` — **commit them**;
  they become permanent regression cases.
- Remember `nil` is a valid payload — include it in seeds.

Fuzzing is not in the per-PR workflow (unbounded time); run it locally or on a schedule.
