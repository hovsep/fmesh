# Contributing to F-Mesh

Contributions of all kinds are welcome: bug reports, docs, examples and code.

## Getting started

1. Check the [existing issues](https://github.com/hovsep/fmesh/issues), or open one to discuss your idea first.
2. Fork the repository and create a feature branch.
3. Open a pull request against `main`.

## Development workflow

F-Mesh needs Go 1.27+. Common tasks are in the Makefile:

```bash
make check   # race + lint + fmt-check: the same gate CI runs
make test    # go test ./...
make fix     # golangci-lint run --fix
make bench   # benchmarks with -benchmem
```

Run **`make check`** before you open a PR. A green `make check` should mean a green CI build.
Linter config is in `.golangci.yml`.

## Code conventions

- **Copy-on-write vs. mutating.** `signal.Signal` and `signal.Group` are copy-on-write: a
  mutating method returns a new value and never touches the receiver. `meta.Meta` and the
  `port`/`component`/`cycle` types mutate in place.
- **Naming follows that split, with no exceptions.** `With*`/`Without*` return a new value (or are
  constructor options). `Set*`/`Add*`/`Remove*` mutate.
- **Metadata on mutating types goes through the store:** `x.Meta().Set(k, v)`. Only the
  copy-on-write types have `WithMeta*`/`WithoutMeta*` methods.
- **`Signal.Payload()` does not fail.** `nil` is a valid payload. Use `sig.As[T]()` or
  `Group.FirstAs[T]()` when you need a type check.
- Fallible methods return `error` last. Infallible transforms (`Filter`, `Map`, `With*`) return
  their type directly.
- The signal **payload** stays `any`: one pipe must carry mixed types. Generics are fine elsewhere
  when they remove real duplication. Avoid `reflect`.
- Priority is **simplicity and a clean API**, not performance.

New code needs tests. Integration suites live in `integration_tests/<topic>/`.

## Documentation

The wiki source is in `docs/wiki/`. It is synced to the GitHub wiki on every push to `main`. Edit
pages there, never in the wiki UI.

Docs are tested: `Example` in `example_test.go` is the README quick start, and `docs_test.go`
checks that every API name in a Go block exists.

## Pull request guidelines

- One feature or fix per PR.
- Say what the change does and why. Link the issue if there is one.
- Update the docs and examples your change affects.

## What will be rejected

- **Any PR offering paid services** (ads, promotional links, offers of commercial products or
  consulting). It is closed without review.
- Drive-by PRs that only bump versions, reword text or reformat code, with no real improvement.

## License

By contributing, you agree to license your work under the [MIT License](LICENSE).
