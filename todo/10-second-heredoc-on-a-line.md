# A second heredoc on one line is lost

`scan_ruby.go:84` finishes a line that opened a heredoc by recursing:

```go
ts2, _ := scanRuby(stateCode, line[i:])
```

The rest of the line is scanned as code, which is right, and the state it
ends in is thrown away, which is not. Two consequences:

- `a = <<~A; b = <<~B` carries the first heredoc and forgets the second,
  so B's body is colored as Ruby and its terminator is a constant.
- `prev` restarts, so a slash in the rest of the line reads as a regex
  where it may be a division.

Neither is common. Both are the kind of thing that shows a reader a line
nobody wrote, which is what `token_test.go` exists to prevent.

## Change

Drop the recursion. Set a pending terminator when the opener is read,
keep scanning the line in the loop already there, and apply the pending
carry on the way out. That keeps `prev` correct for free and makes the
second heredoc the same code path as the first.

Add a table case for both lines above.

## Why after 09

Both changes are in the same twenty lines of the heredoc carry. If `09`
lands, the pending value is a tag rather than an encoded state and this
is a smaller change. If `09` is declined, do this on its own.

## Commit

`ruby: carry a heredoc opened after the first on a line`
