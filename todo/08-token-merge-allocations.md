# The token merge allocates per byte

`tokens.add` merges a run of one class by concatenating:

```go
(*ts)[n-1].text += text
```

Every scanner's `default` branch feeds it one byte at a time, so a line
of punctuation is one allocation per character, each copying everything
merged so far. `token.go:90` says the merge exists so a line of
punctuation is one token rather than twenty, which it achieves by doing
twenty allocations to get there.

## Only if 01 says so

This is the one item on the list where reading the code oversells the
cost. The merged runs are short, the allocator is good at short strings,
and a scanner that is already fast enough does not need this. Do it if
the benchmark from `01` puts `add` in the profile, and drop this file if
it does not.

## Two shapes, if it is worth doing

Hold the trailing run in a builder inside `tokens` and materialize it
when the class changes or the line ends. Contained in `token.go`, and no
scanner changes.

Or store a token as offsets into the line rather than a string, so a
merge moves an index and the emitter slices once. Faster and it changes
`token`, which means every scanner and every test that builds a token by
hand.

Prefer the first. The second is the kind of change that is right and
still not worth what it costs a week into a repo.

## Commit

`token: merge runs without reallocating them`
