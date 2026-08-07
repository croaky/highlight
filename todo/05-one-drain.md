# One way to finish a carry

Eight scanners open with the same six lines. Only the class and the
terminator differ:

- `scan_go.go` (40-57), a backtick and `*/`
- `scan_sql.go` (40-57), a quote and `*/`
- `scan_lua.go` (49-66), `]]` for both
- `scan_javascript.go` (58-76), a backtick and `*/`
- `scan_css.go` (25-33), `*/`
- `scan_json.go` (25-33), `*/`
- `scan_c.go` (42-51), `*/`
- `scan_html.go` (25-33), `-->`

Each one looks for its terminator, emits up to and including it and
returns to code, or emits the rest of the line and returns the carry
unchanged. Written eight times, it is eight chances to get the `+2`
wrong, and `token_test.go` is the only thing that would catch it.

## Change

Add to `token.go`:

```go
// drain emits s up to and including end, or all of s when end is not
// on this line. Returns how much it consumed and whether it closed.
func (ts *tokens) drain(class, s, end string) (int, bool)
```

Each scanner's carry case becomes two lines: call it, advance `i`, and
either clear the state or return.

## Why after 03 and 04

This touches eight files at once. Doing the small moves first means the
diff here is only the drain, and a reviewer can check one helper and
then read eight identical call sites instead of eight rewrites.

## Leave alone

Ruby's `stateRawString` case uses `strings.IndexAny` over five possible
closers, and its heredoc case takes the whole line whatever it holds.
Neither is a search for a fixed terminator. The shell's double quote is
escape-aware, which `scanDelimited` already covers after `03`. HTML's
`stateTag` is a different thing entirely.

## Done when

The eight cases call `drain` and `token_test.go` passes, which is the
check that the byte arithmetic came through.

## Commit

`token: extract the carry drain`
