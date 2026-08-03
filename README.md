# highlight

A small syntax highlighter for server-rendered HTML. One hand-written,
line-oriented scanner per language over a shared token type: Go, sh,
SQL, CSS, JavaScript, Markdown, and hml.

```go
highlight.Code(w, "main.go", src)   // or "go", from a markdown fence
highlight.Diff("main.go", patch)    // a unified-diff patch
```

Five classes come out — `k` keywords, `n` names, `s` strings, `m`
numbers, `c` comments — plus four for a diff's rows: `gi` inserted,
`gd` deleted, `gu` hunk header, `gc` context. Colors are yours; those
nine names are the whole contract.

The curation is the point. It is fast and small because it is
deliberately incomplete and covers only a few languages. A format with
no scanner is escaped and rendered uncolored, so callers can mix. See
`doc.go`.

## This repo is a mirror

Development happens on [cibot](https://dancroak.com/cmd/cibot/), a
self-hosted review and CI server, which holds the branches. GitHub
receives `main` and the tags, so `go get` works and a commit hash is
browsable, and pull requests are closed because there is nothing here to
merge into.

`main` arrives with each merge. A tag is deliberate:

```sh
scripts/tag v0.1.0
```

## License

MIT
