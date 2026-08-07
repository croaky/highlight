# Benchmark the emitters

`doc.go` says the package is fast and small. Nothing here measures
either half, so three of the changes below are guesses until this one
lands.

## Change

Add `BenchmarkCode` and `BenchmarkDiff` to `html_test.go`, both over
the files already in `testdata/`. `Code` gets one sub-benchmark per
sample so a slow scanner is visible on its own. `Diff` gets a patch
built once outside the loop.

Report allocations, since that is what `02` and `08` are about:

```go
b.ReportAllocs()
```

## Why here

`02` claims escaping allocates per token and `08` claims the token
merge allocates per byte. Both are true by reading, but neither is
worth doing if the numbers are small, and neither can be reviewed
without a before and after.

## Done when

`go test -bench=. -benchmem ./...` prints a baseline, and the numbers
are in the commit message so later changes have something to quote.

## Commit

`test: benchmark Code and Diff`
