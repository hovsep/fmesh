# Contributing to F-Mesh

Thanks for your interest in improving F-Mesh! Contributions of all kinds are welcome — bug reports, documentation fixes, examples, and code.

## Getting started

1. Check the [existing issues](https://github.com/hovsep/fmesh/issues) or open a new one to discuss your idea first.
2. Fork the repository and create a feature branch.
3. Make your changes and submit a pull request against `main`.

## Development workflow

F-Mesh requires Go 1.26+. Common tasks are wrapped in the Makefile:

```bash
make check   # race + lint + fmt-check — the same gate CI applies
make test    # go test ./...
make fix     # golangci-lint run --fix
make bench   # benchmarks with -benchmem
```

Before opening a PR, run **`make check`**. It is exactly what CI runs, so a green `make check`
should mean a green build. Linter configuration lives in `.golangci.yml`.

## Code conventions

A few project-wide rules to be aware of:

- **Copy-on-write vs. mutating:** `signal.Signal` and `signal.Group` are copy-on-write — mutating methods return a new value and never touch the receiver. `meta.Labels`/`meta.Scalars` and the `port`/`component`/`cycle` types mutate in place. Naming follows suit, with no exceptions: `With*`/`Without*` return a new value (or are constructor options); `Set*`/`Add*`/`Remove*` mutate.
- **Metadata on mutating types goes through the store:** `x.Labels().Set(k, v)`, not `x.AddLabel(k, v)`. Only the copy-on-write types (`signal.Signal`, `signal.Group`) carry `With*Label`/`With*Scalar` methods, because there they are the only way to produce a modified value.
- **`Signal.Payload()` does not fail.** `nil` is a valid payload; a signal without one means someone skipped `New`. Use `signal.As[T]` when you need the type checked.
- Fallible methods return `error` last; infallible transforms (`Filter`, `Map`, `With*`) return their type directly.
- The signal **payload** stays `any` — one pipe has to carry mixed types. Generics are fine elsewhere where they remove real duplication. Keep `reflect` usage to a minimum.
- Priority is **simplicity and a clean API, not performance**.

New code should come with tests. Integration suites live in `integration_tests/<topic>/`.

## Documentation

The user-facing wiki source lives in `docs/wiki/` and is synced to the GitHub wiki on every push to `main`. Edit pages there — never in the wiki UI directly.

## Pull request guidelines

- Keep PRs focused: one feature or fix per PR.
- Describe what the change does and why; link the related issue if there is one.
- Update documentation and examples affected by your change.

## What will be rejected

- **Any PR offering paid services** — including advertisements, promotional links, or solicitations for commercial products or consulting — will be closed without review.
- Drive-by PRs that only bump versions, reword text, or reformat code with no substantive improvement.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
