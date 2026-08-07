# One way to read a name

Five scanners spell the same loop by hand:

```go
j := i + 1
for j < len(line) && isIdent(line[j]) {
	j++
}
```

`scan_go.go:84`, `scan_c.go:71`, `scan_json.go:56`, `scan_sql.go:83`, and
`scan_sh.go:74` with `|| line[j] == '-'` added. Meanwhile the most
generic name in the package, `identLen`, is in `scan_ruby.go` and is not
generic at all: it takes a trailing `?` or `!` because `empty?` and
`save!` are Ruby method names.

## Change

Add to `token.go`:

```go
// identEnd is the index just past the name starting at line[i].
func identEnd(line string, i int) int
```

Use it in Go, C, JSON, and SQL. Rename Ruby's `identLen` to
`rubyIdentLen` so the general name is free and the special one says what
it is.

Leave the shell and CSS alone. The shell takes a hyphen mid-name so a
flag is one word, and CSS has `scanCSSIdent` for the same reason. Both
are a different question than "is this a name", and folding them in
would mean a parameter that only ever has two values.

## Done when

The four scanners call `identEnd`, `rubyIdentLen` documents the `?` and
`!` rule where it is used, and no table changes.

## Commit

`token: extract identEnd`
