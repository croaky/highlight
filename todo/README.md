# Todo

One file per change, numbered in the order they should land. Each is
meant to be a single cibot change: open it, review it, land it, then
start the next. The numbering is the dependency order, not a priority
list.

The order follows three rules:

- Measurement first. A performance change with no benchmark behind it
  is a guess, so `01` lands before anything that claims to be faster.
- Moves before rewrites. `03` and `04` relocate shared code without
  changing behavior, which makes them cheap to review. Doing them
  first keeps the big diff in `05` down to the one thing it is for.
- Decisions last. `09` and `10` change the carry, and `09` needs
  agreement on the shape of `state` before anybody writes it.

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
