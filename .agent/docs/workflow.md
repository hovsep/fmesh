# Workflow

## Renames

- Don't rely solely on sed — it misses bare names, function names, strings, comments
- Run `go build ./...` after each rename wave before moving on
- When updating callers, check if other packages have their own method with the same name and different semantics — don't touch them without user approval

## Edits

- After editing a source file, verify it wasn't accidentally truncated
- Prefer small targeted edits over large multi-line replacements

Constraint/pattern rules are in `CLAUDE.md` (hard rules); comment style is in
`design.md` (comment hygiene).
