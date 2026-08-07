# The two emitters have different shapes

```go
func Code(w io.Writer, name, src string) error
func Diff(filename, patch string) template.HTML
```

`doc.go` says HTML is one consumer rather than the interface, and that
another emitter is a new file with no change to any scanner. That holds
for the token layer and not for this: a second emitter has to pick one of
these two shapes, and there is no reason in the package for which.

A caller in a template wants `Diff`'s shape for both. A caller streaming
a file to a response wants `Code`'s shape for both.

## Not a change yet

Nothing depends on both shapes today, so this is a decision to make
before something does, not work to schedule. Three ways out:

- Both return `template.HTML`. Simplest to call from a template, and it
  buffers a whole file in memory, which is what `Code` already does at
  `html.go:20`.
- Both take an `io.Writer`. Streams, and every template caller writes a
  wrapper.
- Keep the pair and say why in `doc.go`: `Code` is for a file and `Diff`
  is for a fragment. That is a real distinction if we mean it.

The third is the cheapest and the only one that needs writing rather than
deciding. Whichever we pick, it goes in `doc.go` next to the two-layer
paragraph, because that paragraph is what makes the reader expect a rule
here.
