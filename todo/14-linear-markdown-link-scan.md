# Make the Markdown link-text scan linear

Title: `security: stop repeating the Markdown bracket search`

Finding: `find_01RRBEcZEubE5KoCMMv4kWK5` (medium): quadratic time in
`scanMarkdownInline` on a line of unclosed brackets.

Outcome: fix. The fix is two lines and changes no output.

## Threat model

The attacker, the boundary and the consequence are the same as in
`todo/13-linear-regex-and-heredoc-scan.md`. The input must have a
Markdown name (`md`, `markdown`, `.md` or `.markdown`). One line of
about 1 MB of `[` costs about 5×10¹¹ byte compares.

## Why this sits after 13

It touches only `scan_md.go`, so it does not depend on 13. It sits
after 13 because 13 has the high finding. If 13 adds the long-line
test, this change adds a Markdown case to that test.

## Cause

At each `[`, `scanMarkdownInline` calls `strings.IndexByte(s[i:], ']')`.
When the line has no `]`, each call reads the rest of the line and
returns -1. The scanner moves one byte and repeats the call.

## Options

1. Remember the miss (recommended). After one search finds no `]`,
   set a local bool and skip the search for the rest of the line. No
   `]` after position i means no `]` after any later position, so the
   output does not change for any input. Cost: one bool and one
   condition.
2. Remove link-text coloring. Delete the `[` branch, and `[text]`
   prints with no class. Cost: a lost feature, for about eight fewer
   lines. The fix in option 1 is smaller than the loss, so do not
   remove it.

## Tests

The `scan_md_test.go` table and `testdata/sample.md` must not change.
If 13 keeps the long-line test, add a 512 KB line of `[` under a
Markdown name.

## Not in this change

The backtick branch in the same function returns after a miss, so it
is already linear.
