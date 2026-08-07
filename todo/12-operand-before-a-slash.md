# A literal does not count as an operand before a slash

A slash divides where an operand just ended and opens a regex where a
value belongs. Both scanners decide with `endsOperand(prev)`, where
`prev` is the byte the previous token *started* with. A literal starts
with its delimiter, so a literal never ends an operand:

```
ruby  x = "a" / b / c        ->  :x = |s:"a"|: |s:/ b /|: c
ruby  x = %w[a] / b / c      ->  :x = |s:%w[a]|: |s:/ b /|: c
ruby  x = <<~SQL / b / c     ->  :x = |s:<<~SQL|: |s:/ b /|: c
js    x = `a` / b / c        ->  :x = |s:`a`|: |s:/ b /|: c
js    x = "a" / b / c        ->  :x = |s:"a"|: |s:/ b /|: c
```

`/ b /` is colored as a regex in every one. The control case is right:
`x = n / b / c` divides, because a name's first byte is also its last.

`scan_javascript.go:55` is the same bug with a comment claiming the
opposite. It sets `prev = '`'` after a closed template literal and says
"a template literal is a value, so a slash after it divides", but
`endsOperand('`')` is false, so it does not.

## Change

`prev` is the wrong question. Replace it with whether the last token
ended an operand, which the branch that emitted the token knows and the
byte cannot recover: true after a name, a number, a string, a regex, a
percent literal, a heredoc opener, and a closing bracket; false after an
operator or an open bracket.

`endsOperand` then has one caller shape left and folds into the places
that set the flag. Both scanners have a single `prev = c` at the bottom
of the loop, so this moves that assignment into the branches.

## Both scanners

`endsOperand` is shared by Ruby and JavaScript, and it is the same rule
in both, which is why it sits in `token.go`. Fixing one and not the
other would leave the shared helper serving two meanings.

## Commit

`token: count a literal as an operand before a slash`
