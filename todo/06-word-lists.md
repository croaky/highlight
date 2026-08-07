# Keyword maps that read as word lists

Eight files hold a `map[string]bool` literal where a third of the
characters are `: true`. `goKeywords`, `goNames`, `cKeywords`,
`jsKeywords`, `jsNames`, `luaKeywords`, `luaNames`, `rubyKeywords`,
`shKeywords`, `sqlKeywords`, `jsonWords`.

A keyword list is the part of a scanner a reader is most likely to edit,
and right now it is the part that is hardest to read.

## Change

One constructor in `token.go`:

```go
// words is a set of the words given, for the keyword and name lists
// each scanner matches against.
func words(w ...string) map[string]bool
```

Each list becomes the words and nothing else:

```go
var goKeywords = words(
	"break", "case", "chan", "const",
	...
)
```

## Not doing yet

A generated `switch word { case "break", ...: return true }` would beat
a map: no hashing, no map to build at init, and the compiler switches on
length first. That is worth doing only if `01` shows keyword lookup in
the profile, and it is a separate change either way, because a generator
is a thing to maintain and this is not.

## Done when

Every list is a call to `words`, the comments above them are unchanged,
and no test moves.

## Commit

`token: spell the keyword lists as word lists`
