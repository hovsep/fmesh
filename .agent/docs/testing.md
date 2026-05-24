# Testing

## Style

- Tests live alongside source, same package (`package signal` not `package signal_test`)
- Table-driven by default; `t.Run` subtests for grouped inline assertions
- `require` for preconditions and error checks (stops on failure); `assert` for value checks (continues)
- No assertion helpers — use plain `assert`/`require` directly
- Only allowed helper: `mustXxx()` panic-on-error for fixture setup, never for assertions
- Use `assert.InDelta` for float64 comparisons (tolerance `1e-9` for exact values, larger for computed averages)

## What to cover

- CoW invariant: verify receiver is unchanged after every mutating method on `signal.Signal` and `signal.Group`
- Edge cases: nil payload, empty group/collection, missing scalar name
- `meta.Scalars`: `Min`/`Max` return `ok=false` on empty store; `Average` returns `ok=false` on empty store; `Sum` of empty = 0
- Cross-entity aggregation on groups: `AvgScalar`/`MinScalar`/`MaxScalar` return `ok=false` when no element has the named scalar; `SumScalar` returns 0
- Group metadata separation: group's own Labels/Scalars must not bleed into element Labels/Scalars and vice versa
- `signal.Group` batch methods (`WithLabelOnEach`, `WithScalarOnEach`, etc.) must preserve the group's own metadata on the returned group
