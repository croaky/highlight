# Name the heredoc tag instead of counting it

A Ruby heredoc's terminator is encoded as arithmetic on the state:
`stateHeredoc + 1 + state(i)`, where `i` indexes `heredocTags`. The
comment at `scan_ruby.go:234` explains it well, and it is still the
hardest thing in the package to read. Three functions exist only to
maintain the encoding: `heredocState`, the `st > stateHeredoc` branch in
`heredocEnds`, and the `heredocTags` list itself.

It also caps exactly which heredocs are scanned correctly at nineteen
tags. A `<<~MIGRATION` falls back to the loose rule and ends at the first
line that is all capitals, which is a body a reader can lose.

## Decision needed first

Do not write this until we agree on the shape. The proposal:

```go
type state struct {
	kind kind
	tag  string // the heredoc terminator, when kind is heredoc
}
```

`state` stays comparable, so `st == stateCode` still works, the zero
value still means code, and `html.go` does not change at all. Any
terminator then works exactly, and `heredocTags`, `heredocState`, and
half of `heredocEnds` are deleted.

The cost is every test that names a state literal. `stateBlockComment`
becomes `state{kind: blockComment}` in a dozen tables, which is a wide
diff of mechanical edits, and mechanical edits in a table test are where
a wrong expectation hides.

The alternative is to leave the arithmetic and add tags to the list as
they come up. That is cheap forever and wrong for anything not listed.

## Why last

It changes the type every scanner signs, so it wants the shared helpers
already extracted and every scanner already passing. It is also the only
item here that is a judgment call rather than a cleanup.

## Commit

`token: carry the heredoc tag in the state`
