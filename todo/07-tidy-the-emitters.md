# Tidy the emitters

Three small things in `html.go`, all in one change because they are all
the same file and none is worth a review of its own.

## Change

`Code` splits with `strings.Split` at `:22` and `Diff` splits with
`strings.SplitSeq` at `:52`. Use `SplitSeq` in both. `Code` does not
need the line slice and does not keep it, so the allocation is pure
waste, and two spellings of one thing in one file is a question a reader
has to answer twice.

Grow the buffer before the loop. The output is the source plus the spans,
so `len(src)` is the floor and something like `len(src)+len(src)/4` saves
the regrowth on a file of any size.

Merge the `'@'` and `'\\'` cases at `:60` and `:65`. The bodies are
identical apart from the state reset, which is what the `@` case is
actually for.

## Done when

`html_test.go` and `diff_test.go` pass unchanged.

## Commit

`html: one way to split a source into lines`
