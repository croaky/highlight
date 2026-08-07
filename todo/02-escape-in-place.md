# Escape into the buffer

Every write path calls `template.HTMLEscapeString`, which builds a new
string for each token whether or not the token holds a byte worth
escaping. Most tokens do not: a keyword, a name, a number, and the
whitespace between them are all returned unchanged after a full scan
and an allocation.

Four call sites, `html.go:63`, `:67`, `:106`, and `:120`/`:126`.

## Change

One function in `html.go`:

```go
// writeEscaped writes s with the bytes that could become markup
// replaced, and the runs between them written as they are.
func writeEscaped(buf *bytes.Buffer, s string)
```

It walks s for the escapable bytes, writes each untouched run with one
`WriteString`, and writes the replacement for each byte it finds. A
token with nothing to escape becomes a single write and no allocation.

Escape exactly what `template` escapes: `'`, `"`, `&`, `<`, `>`, and
NUL. Anything narrower changes what the package promises in the
Escaping section of `doc.go`.

## Watch for

The replacements have to match byte for byte, `&#39;` and `&#34;` and
not `&apos;` and `&quot;`, or every test that spells out expected HTML
fails and the diff hides whether the escaping is still right. Copy the
table from `html/template`'s source rather than writing it from memory.

## Done when

`token_test.go` and the emitter tests pass unchanged, and the `01`
numbers show fewer allocations per operation.

## Commit

`html: escape into the buffer instead of into a new string`
