# Move the shared helpers into token.go

`AGENTS.md` says a scanner is one file per language and `token.go` holds
what they share. Four helpers are in the wrong file, and one of them is
written three times.

- `scanRegex` and `endsOperand` live in `scan_javascript.go` and are
  called from `scan_ruby.go`. That is one language scanner reaching into
  another's, which is the arrangement the layout exists to avoid.
- `endsRubyOperand` is `endsOperand` with `isIdent` in place of
  `isJSIdent`. The only difference is `$`, which Ruby handles in its own
  sigil branch before the slash branch is reached.
- `isSpace` is in `scan_sh.go`, while `scanCDirective` and
  `jsonStringClass` spell out `' ' || '\t'` inline.
- `scanQuoted` in `token.go`, `scanDelimited` in `scan_ruby.go`, and
  `scanShellDouble` in `scan_sh.go` are the same escape-aware scan to a
  closing byte.

## Change

Move `scanRegex`, `endsOperand`, and `isSpace` to `token.go`. Delete
`endsRubyOperand` and point Ruby at `endsOperand`. Use `isSpace` in
`scan_c.go` and `scan_json.go`.

Keep one delimiter scan:

```go
// scanDelimited returns the length of the run up to and including
// close, and whether close was found.
func scanDelimited(s string, close byte) (int, bool)
```

`scanQuoted` becomes a wrapper that starts past the opening quote and
drops the bool. `scanShellDouble` goes away, and the shell's
double-quote branch calls `scanDelimited` directly.

## Watch for

The offset convention is the whole risk. `scanQuoted` starts its loop at
1 because `s[0]` is the opening quote, and `scanDelimited` starts at 0
because Ruby already consumed the `%q` and the delimiter. Get that
wrong and a two-character string swallows the rest of the line. The
Ruby and shell tables cover it, but check the returned lengths by hand
before trusting them.

## Done when

No scanner imports another scanner's helper, and the tests pass with no
table edited.

## Commit

`token: move the shared helpers out of the language files`
