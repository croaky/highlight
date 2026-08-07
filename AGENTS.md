# Agents guide

highlight colors source code for server-rendered HTML. See `README.md`
for why it exists and `doc.go` for the scope and the class names.

## Architecture

One package at the repo root, standard library only. Two layers, and the
split is the design:

- `token.go` — the token type, the carry state, the `scanFunc`
  signature, the shared byte helpers, and `scannerFor`. Knows nothing
  about output.
- `scan_*.go` — one scanner per language, one file each. A scanner reads
  a line and emits tokens; it never writes markup. The suffix is spelled
  out when the short name is a GOOS or GOARCH: `scan_javascript.go`,
  because Go would build `scan_js.go` only for js/wasm.
- `html.go` — the exported surface, `Code` and `Diff`, and the escaping.
  HTML is one consumer, not the interface: another emitter is a new file
  here and no change to any scanner.

## Checks

The root `Checkfile` is the list, and CI runs it on every push. Run the
same things before committing:

```sh
goimports -local "$(go list -m)" -w .
go vet ./...
go test -race -cover ./...
git ls-files -z '*.go' | xargs -0 gopls check -severity=hint
```

The package imports nothing outside the standard library. Tests import
`github.com/croaky/is` for assertions, and that is the only dependency.
Taking another is a design decision, not a step.

## Tests

A scanner's test is a table of one line each, in and out states
included, since the carry is what the design is for. `token_test.go`
scans `testdata/sample.*` and checks the tokens still spell the file: a
scanner that drops or repeats a byte shows a reader a line nobody wrote.

A new language is a `scan_*.go`, a case in `scannerFor`, a table test,
and a `testdata/sample.*` that exercises whatever it carries.

Assertions come from `github.com/croaky/is`: `is := is.New(t)`, or
`is.NewRelaxed(t)` where a case should report every mismatch rather
than stop at the first. Pick the helper that names the check, `Eq` and
`NotEq` by default, `True` only for a predicate with no want to name.
Arguments are (got, want), and `Eq` takes `any`, so type the literal
rather than converting the value under test.

Incompleteness is allowed and does not need a test. Wrong color does.

## Todo

`todo/` is the work that is planned and not done, one file per change,
numbered in the order the changes should land. A file says why its
change sits where it does and what it is deliberately not doing, which
is the part that would otherwise be lost. `todo/README.md` is the order
and the rules behind it.

A `todo:` commit is one that changes the plan: a file added, reordered,
or dropped as no longer worth doing. Doing the work is not a `todo:`
commit. It takes the prefix of what it acts on, and it deletes the file
it finishes, in the same commit as the code. A file that outlives its
change becomes a note about something already fixed, and a directory of
those is one nobody opens.

Two things fall out of a deletion. The numbers are positions and not
names, so a gap is fine and nothing gets renumbered to close it. And a
file that pointed at the deleted one has to say what it means instead,
usually by naming the thing that now exists.

A file's premise can turn out to be wrong, since it was written before
the work. Say so in the commit message rather than quietly doing
something else, and fix the files that assumed the same thing.

## Changes

Work happens on a cibot change. `cibot checkout` allocates one and
prints a worktree; `cibot edit` sets its title and description. Do the
edit before the code, not after. A change with neither is a blank row on
the dashboard and a blank `cibot show`, so nobody looking at either can
tell what it is or whether it overlaps what they are about to start. A
rough sentence beats an empty one, and the description gets rewritten
before the merge anyway.

## Commits

- Prefix with what the change acts on: the language (`go:`, `css:`,
  `sql:`), or the layer (`token:`, `html:`, `doc:`, `todo:`, `ci:`).
  Not `highlight:` — every commit here is highlight.
- Imperative mood, lowercase except proper nouns. Hard-wrap at 72.
- Include _why_, not just _what_. See `git log` for examples.
- Sign your work with a `Co-Authored-By` trailer.

## Releases

cibot is origin and holds no tags. `scripts/tag vX.Y.Z` publishes one
annotated tag to the GitHub mirror, which is what a `go get` resolves.
