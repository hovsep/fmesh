# Testing

## Style

- Tests live beside the source, in the same package (`package signal`, not `package signal_test`).
  Godoc examples (`example_test.go`) and `docs_test.go` are the exception.
- Table-driven by default; `t.Run` subtests for grouped inline assertions.
- `require` for preconditions and error checks (stops); `assert` for value checks (continues).
- No assertion helpers — use `assert`/`require` directly.
- The only allowed helper is `mustXxx()`, panic-on-error, for fixture setup — never for assertions.
  Shared ones live in `internal/testutil` (usable from `integration_tests/` and other external
  test packages). In-package tests of `fmesh` and `port` keep local copies (import cycle).
- `assert.InDelta` for float64 (tolerance `1e-9` for exact values, larger for computed averages).
- Comments say why a case exists (the bug it pins), never what the assertion does. One or two
  lines — the same brevity rule as source ([design.md](design.md)).

## What to cover

- **CoW**: the receiver is unchanged after every mutating method on `signal.Signal` and
  `signal.Group`.
- **Edge cases**: nil payload, empty group/collection, missing metadata key, a key holding the
  other type.
- **Group metadata separation**: a group's own `Meta()` must not bleed into its elements' and vice
  versa. `signal.Group.Meta()` returns a clone — mutating it must not change the group.
- **`signal.Group` batch methods** (`WithMetaOnEach`, `WithoutMetaOnEach`) keep the group's own
  metadata on the result.
- **Anything taking a port name as a string**: cover a name that matches no port. An unknown name
  turns into an empty collection (vacuously ready) or a nil port (a panic on first use), so a
  passing test proves nothing unless it uses a missing port.
- **Typed payload accessors**: a wrong payload type, a nil payload and a nil signal all return an
  error or the default — never panic.

## Coverage is not a usage signal

A method used only by its own unit test is not dead. The library's callers are in
`fmesh-examples` and `fmesh-graphviz`, which no test here mentions. Never delete exported API on
"own-package test only" grounds; see [downstream.md](downstream.md).

## Documentation is tested too

- `Example` in `example_test.go` is the README quick start and runs in CI. Keep the two in sync —
  change one, change the other.
- `docs_test.go` scans the ```go blocks in `README.md`, `CONTRIBUTING.md`, `docs/wiki/*.md` and the
  plugin READMEs (`plugin/README.md`, `plugin/*/README.md`):
  - `TestDocs_ReferenceOnlyExistingAPI` — every qualified reference (`component.X`, `signal.X`, …)
    names a real exported symbol or method. Also scans `CHANGELOG.md`.
  - `TestDocs_NoRemovedMethodNames` — rejects a denylist of removed method names, which catches
    calls on a variable (`c.Meta().AddLabel(...)`). Also scans Go source comments.
  - `TestDocs_MethodCallsExistSomewhere` — every `.SomeExported(` call names a method or function
    defined in this module, or is on a small allowlist of external names (stdlib, profiler).
- Not checked: compilation (snippets are fragments with elisions), argument counts, and prose
  outside Go blocks. Method names count as valid for a package qualifier, because docs shadow
  package names with variables (`port.Signals()` on a `*Port` named `port`).
- `.agent/docs/` is not scanned. Check its snippets against the code by hand.
