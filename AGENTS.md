# Agents guide

highlight colors source code for server-rendered HTML. See `README.md`
for why it exists and `doc.go` for the scope and the class names.

## Writing

Write every word in ASD-STE100 Simplified Technical English (STE):
Markdown docs, code comments, commit messages, and replies in an agent
conversation. See
<https://en.wikipedia.org/wiki/Simplified_Technical_English>.

STE is a controlled English for technical writing: one meaning per
word, one idea per sentence, and the actor named. It is not a house
style. It exists so a reader who is tired, or reading a second
language, or an agent matching on words, all read the same sentence the
same way.

- One idea per sentence. Keep an instruction to 20 words and a
  description to 25.
- Active voice, present tense, and the actor named: say what acts,
  rather than writing "the token is refused".
- One word, one meaning. Keep a term the same everywhere rather than
  varying it for tone.
- Use the simple verb, not a noun made from it: "run the formatter",
  not "perform execution of the formatter".
- Cut what carries nothing: "simply", "just", "note that", "in order
  to".
- Put a warning or a limit before the step it applies to.

Apply it to prose, not to code: an identifier, a command, and a quoted
error message stay as they are.

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

The root `Checkfile` is the list, and CI runs it on every push. Read it
there rather than here, and run it before committing: a second copy of
the list is one that drifts.

Two of its entries check rather than write, since a CI job that
rewrote source would have nowhere to put it. Run the writing form
first, and the check passes:

```sh
goimports -local "$(go list -m)" -w .
dprint fmt
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

After a push, read the checks with `git push && cibot show --wait`
rather than sleeping and then reading. The farmer holds the request open
and answers within a second of the last check, so a sleep is either time
spent waiting for an answer that already arrived or too short to reach
one. Too short is the worse half: a `cibot show` that lands before the
push is recorded reports the previous commit's checks, green, about the
wrong code. `--wait` follows the commit in the worktree it runs in,
exits nonzero when a check failed, and gives up after ten minutes
(`--timeout`).

Write the title and description as the commit message they become. A
merge is a squash that takes them verbatim, so a title like "Escape into
the buffer" lands in `git log` next to `html: escape into the buffer`,
and the log stops being one thing. Title is the subject, description is
the body, both under the rules below. The first pass can be rough; make
it a real subject line before the merge.

## Commits

- Subject: `scope: imperative summary`, at most 50 characters, lowercase
  except proper nouns, no trailing period.
- The scope is what the change acts on: the language (`go:`, `css:`,
  `sql:`), or the layer (`token:`, `html:`, `doc:`, `todo:`, `ci:`).
  Not `highlight:` — every commit here is highlight.
- Blank line, then a body hard-wrapped at 72, in sentences with
  punctuation.
- Include _why_, not just _what_, and cut whatever is inessential. See
  `git log` for the scopes in use and for examples.
- Sign your work with a `Co-Authored-By` trailer.

## Releases

cibot is origin and holds no tags. `scripts/tag vX.Y.Z` publishes one
annotated tag to the GitHub mirror, which is what a `go get` resolves.
