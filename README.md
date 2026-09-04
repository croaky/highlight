# highlight

A small syntax highlighter for server-rendered HTML. One line-oriented scanner
per language: Go, SQL, CSS, JavaScript, C, Ruby, JSON, Lua, HTML, Markdown,
shell, and [hml](https://github.com/croaky/hml).

```go
highlight.Code(w, "main.go", src)   // or "go", from a markdown fence
highlight.Diff("main.go", patch)    // a unified-diff patch
```

That generates CSS classes `k` keywords, `n` names, `s` strings, `m`
numbers, `c` comments and a few for a diff's rows: `gi` inserted,
`gd` deleted, `gu` hunk header, `gc` context. `Diff` wraps the whole
patch in one `diff` span. Colors are yours to style. Those names are
the contract.

It is fast and small because it is deliberately incomplete and covers only a few
curated languages. A format with no scanner is escaped and rendered uncolored,
so callers can mix. See `doc.go`.

## GitHub repo is a mirror

Development happens on [cibot](https://dancroak.com/cmd/cibot/), a
self-hosted review and CI server, which holds in progress branches.
GitHub receives `main` and the tags so `go get` works.

## License

MIT
