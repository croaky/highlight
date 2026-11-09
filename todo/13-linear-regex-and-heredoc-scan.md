# Make the regex look-ahead and the heredoc carry linear

Title: `security: bound regex look-ahead and Ruby heredoc tags`

Findings:

- `find_0113aMQGMW5caAt5k7BzcihW` (high): quadratic regex look-ahead
  in `scanJS` and `scanRuby`.
- `find_01R3RsZfPCuRT2KCKDB2yYmo` (medium): an unbounded Ruby heredoc
  tag in the carry makes `Diff` cost rows times tag length.

Outcome for both: fix. Each fix is a few lines, adds no moving part,
and makes no scan slower.

## Threat model

The attacker is any user who can send text to an app that passes it
to `Code` or `Diff`. For an internal app on the public internet, that
is any user with an account, or anyone if the upload path has no
login. The boundary is the call into `Code` or `Diff`: the library
puts no limit on input size or on work per line. The consequence is
CPU time. One line of about 1 MB holds one core for minutes, and a few
requests stop the host. No data leaks and no output changes.

Accepting the risk is not the better trade here. Each consumer would
need its own size cap, and the fix in the library is smaller than one
cap.

## Why these two are one change

Both findings edit `scan_ruby.go`, and the regex fix also edits
`scanRegex` in `token.go`. One change gives one review of the Ruby
scanner and one table update.

## Regex look-ahead

Cause: `scanJS` and `scanRuby` call `scanRegex(line[i:])` at each slash
that follows no operand. A failed call reads to the end of the line.
The next slash starts a new full read. A line of N bytes of `/[` or
`/\` costs about N²/4 steps.

Options:

1. Per-line flag (recommended). After one `scanRegex` call fails on a
   line, the scanner reads each later slash on that line as division.
   Cost: one bool and one condition in each of the two scanners. Work
   per line becomes linear. Color changes only on a line that has an
   unclosed regex and then a closed one, such as `/[ /x/`. The current
   answer for that line is a guess too.
2. Cap the look-ahead at a constant, for example 256 bytes. Cost: one
   constant in `scanRegex`. Work stays O(N times the cap). A regex
   longer than the cap loses its color.
3. Make `scanRegex` resume from the state of the failed read. Exact,
   but it adds state to a shared helper for a case no real code has.
   Do not do this.

Do not remove regex detection. Without it, a quote in `/"/` opens a
string and colors the rest of the line wrong.

## Heredoc tag

Cause: `heredocTagLen` accepts a tag of any length. `queueHeredoc`
puts the tag in `state.tag` or `state.next`. `Diff` compares `oldSt`
and `newSt` on each context row. Two equal-length tags that differ in
the last byte make each compare read the full tag.

Fix: `heredocTagLen` returns 0 for a tag longer than 64 bytes. The
scanner then reads `<<` as the append operator. Cost: one constant and
one condition. Real tags such as `SQL`, `EOS` and `MIGRATION` are far
below the limit.

Do not make the state compare constant time with a hash or an ID. That
adds machinery to `state`, and `state` must stay comparable with `==`.

## Tests

- Add a row to the `scanRuby` table: a 65-byte tag is not a heredoc,
  and the carry stays `stateCode`. This test is exact and costs
  nothing.
- Add rows to the `scanJS` and `scanRuby` tables for `/[ /x/`, so the
  table records the color that option 1 picks.
- A time test cannot fail on a threshold without flakes. One choice is
  a test that scans a 512 KB line of `/[` under a JavaScript name and
  a Ruby name. Linear, it takes milliseconds. Quadratic, it takes
  minutes and the `test` check shows the regression. Measure the cost
  under `-race` before you keep it. If it adds more than 100 ms to the
  `test` check, drop it and keep a code comment on the bound.

## Not in this change

The Markdown bracket scan is `todo/14-linear-markdown-link-scan.md`.
It touches other code. The other `strings.IndexByte` callers in
`scan_c.go`, `scan_hml.go`, `scan_html.go` and `scan_sh.go` move past
what they read, so they are already linear.
