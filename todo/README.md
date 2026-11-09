# Todo

The order:

1. `13-linear-regex-and-heredoc-scan.md` fixes the high finding in the
   JavaScript and Ruby scanners and a medium finding in the Ruby
   scanner.
2. `14-linear-markdown-link-scan.md` fixes a medium finding in the
   Markdown scanner. It does not depend on 13.

Both plans fix work per line that grows with the square of the line
length. A security scan found these issues. Each plan names its
finding IDs. After the change lands, mark each finding as fixed by
that commit.

## Threat model

Judge a security finding by these three parts. A severity label alone
is not enough.

- Attacker: any user who can send text to an app that calls `Code` or
  `Diff`. For an internal app on the public internet, that is any user
  with an account, or anyone if the upload path has no login.
- Boundary: the call into `Code` or `Diff`. The library puts no limit
  on input size or on work per line, and it does no I/O.
- Consequence: CPU time in the host process. The output is escaped, so
  a scanner bug is wrong color, not injected markup.

A fix in the library is better than a size cap in each app, when the
fix is small.

## Rules

`AGENTS.md` holds the rules for this directory: one file per change,
numbered in the dependency order the changes should land in, and
deleted by the commit that finishes it. Each file is one sockeye change:
open it, review it, land it, then start the next. The next file starts
past the highest number already used.

Every change runs the `Checkfile` list before it opens. `token_test.go`
is the check that matters for most of these: the scanners still have to
spell the sample files byte for byte. A refactor that changes coloring
will show up there or in a scanner's table.
