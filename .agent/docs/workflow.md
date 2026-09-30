# Workflow

## Renames

- Do not rely on sed alone — it misses bare names, function names, strings and comments.
- Run `go build ./...` after each rename wave before moving on.
- When updating callers, check whether another package has its own method with the same name and
  different semantics. Do not touch those without user approval.

## Edits

- After editing a source file, check it was not accidentally truncated.
- Prefer small targeted edits over large multi-line replacements.

Hard rules (constraints, patterns) are in `CLAUDE.md`; comment style is in
[design.md](design.md) ("Comment hygiene").
