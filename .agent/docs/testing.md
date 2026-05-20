# Testing

## Style

- Tests live alongside source, same package (`package signal` not `package signal_test`)
- Table-driven by default; `t.Run` subtests for grouped inline assertions
- `require` for preconditions and error checks (stops on failure); `assert` for value checks (continues)
- No assertion helpers — use plain `assert`/`require` directly
- Only allowed helper: `mustXxx()` panic-on-error for fixture setup, never for assertions

## What to cover

- CoW invariant: verify receiver is unchanged after every mutating method
- Chainable error: verify methods are no-ops when `HasChainableErr()` is true
- Edge cases: nil payload, empty group/collection, error-carrying input
