# Todo

The order:

1. `14-linear-markdown-link-scan.md` fixes a security finding in the
   Markdown scanner: work per line that grows with the square of the
   line length.

`AGENTS.md` holds the rules for this directory: one file per change,
numbered in the dependency order the changes should land in, and
deleted by the commit that finishes it. Each file is one sockeye change:
open it, review it, land it, then start the next. The next file starts
past the highest number already used.

Every change runs the `Checkfile` list before it opens. `token_test.go`
is the check that matters for most of these: the scanners still have to
spell the sample files byte for byte. A refactor that changes coloring
will show up there or in a scanner's table.
