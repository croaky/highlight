# Todo

One file per change, numbered in the order they should land. Each is
meant to be a single cibot change: open it, review it, land it, then
start the next. The numbering is the dependency order, not a priority
list.

One rule is left, now that the moves and the measured changes have
landed:

- Decisions last. `11` is a decision with nothing scheduled behind it,
  and `12` is a coloring bug `10` turned up while reading the carry.

Every change runs the `Checkfile` list before it opens:

```sh
goimports -local "$(go list -m)" -w .
go vet ./...
go test -race -cover ./...
git ls-files -z '*.go' | xargs -0 gopls check -severity=hint
```

`token_test.go` is the check that matters for most of these: the
scanners still have to spell the sample files byte for byte. A refactor
that changes coloring will show up there or in a scanner's table.

Delete a file when its change lands.
