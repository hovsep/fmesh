# Testing

## Style

- Tests live alongside source, same package (`package signal` not `package signal_test`)
- Table-driven by default; `t.Run` subtests for grouped inline assertions
- `require` for preconditions and error checks (stops on failure); `assert` for value checks (continues)
- No assertion helpers — use plain `assert`/`require` directly
- Only allowed helper: `mustXxx()` panic-on-error for fixture setup, never for assertions.
  Shared implementations live in `internal/testutil` (usable from `integration_tests/` and other
  external test packages); in-package tests of `fmesh` and `port` keep local copies (import cycle)
- Use `assert.InDelta` for float64 comparisons (tolerance `1e-9` for exact values, larger for computed averages)
- Comments explain why a case exists (the bug it pins down), never what the assertion does — one or two lines, same brevity rule as source (see [design.md](design.md))

## What to cover

- CoW invariant: verify receiver is unchanged after every mutating method on `signal.Signal` and `signal.Group`
- Edge cases: nil payload, empty group/collection, missing scalar name
- `meta.Scalars`: `Min`/`Max` return `ok=false` on empty store; `Average` returns `ok=false` on empty store; `Sum` of empty = 0
- Cross-entity aggregation on `signal.Group`: `AvgScalar`/`MinScalar`/`MaxScalar` return `signal.ErrScalarNotFoundInGroup` when no element has the named scalar; `SumScalar` returns 0
- Group metadata separation: group's own Labels/Scalars must not bleed into element Labels/Scalars and vice versa
- `signal.Group` batch methods (`WithLabelOnEach`, `WithScalarOnEach`, etc.) must preserve the group's own metadata on the returned group
- Anything taking a port name as a string: cover the name that resolves to no port. An unresolved name reaches the assertion as an empty collection (vacuously ready) or a nil port (a panic at the first dereference), so the passing test proves nothing unless it names a port that does not exist
- Typed payload accessors: a wrong payload type, a nil payload, and a nil signal must all return an error or the default — never panic

## Coverage is not a usage signal

A method exercised only by its own unit test is not thereby dead — the library's callers are in
`fmesh-examples` and `fmesh-graphviz`, and no test in this repo will ever mention them. Do not use
"own-package test only" as grounds for deleting exported API; see [downstream.md](downstream.md).

## Documentation is tested too

- `Example` in `example_test.go` is the README quick start. It runs in CI, so the README's headline
  snippet cannot rot. Keep the two in sync — if you change one, change the other.
- `TestDocs_ReferenceOnlyExistingAPI` (`docs_test.go`) parses every ```go block in `README.md`,
  `CHANGELOG.md`, `CONTRIBUTING.md` and `docs/wiki/*.md` and asserts that every qualified reference
  to this project's API (`component.X`, `signal.X`, …) is a symbol that actually exists. It is how
  a rename gets caught in the docs rather than by a user copying a dead snippet.

  It deliberately does not compile the wiki snippets: they are fragments with elisions, so wrapping
  them into buildable files is guesswork. It also accepts method names as valid for a package
  qualifier, because docs shadow package names with variables (`port.Signals()` is a `*Port` called
  `port`). Argument counts are not checked.
