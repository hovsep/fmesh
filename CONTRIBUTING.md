# Contributing to F-Mesh

Thanks for your interest in improving F-Mesh! Contributions of all kinds are welcome — bug reports, documentation fixes, examples, and code.

## Getting started

1. Check the [existing issues](https://github.com/hovsep/fmesh/issues) or open a new one to discuss your idea first.
2. Fork the repository and create a feature branch.
3. Make your changes and submit a pull request against `main`.

## Development workflow

F-Mesh requires Go 1.26+. Common tasks are wrapped in the Makefile:

```bash
make test    # go test ./...
make race    # go test -race ./...  (run this for scheduler/port changes — concurrency is core)
make lint    # golangci-lint run ./...
make fmt     # go fmt ./...
make check   # race + lint
make bench   # benchmarks with -benchmem
```

Before opening a PR, please make sure `make test`, `make lint`, and `make fmt` are clean. Linter configuration lives in `.golangci.yml`.

## Code conventions

A few project-wide rules to be aware of:

- **Copy-on-write vs. mutating:** `signal.Signal` and `signal.Group` are copy-on-write — mutating methods return a new value and never touch the receiver. `meta.Labels`/`meta.Scalars` and the `port`/`component`/`cycle` types mutate in place. Naming follows suit: `With*`/`Without*` return a new value; `Set*`/`Add*`/`Remove*` mutate. Don't mix them.
- Fallible methods return `error` last; infallible transforms (`Filter`, `Map`, `With*`) return their type directly.
- No generics in the core API, and keep `reflect` usage to a minimum.
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
