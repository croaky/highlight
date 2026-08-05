package highlight

import (
	"testing"

	"github.com/croaky/is"
)

func TestScanSQL(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "keywords are case-insensitive",
			line: "select id from jobs WHERE n = 1",
			want: "k:select|: id |k:from|: jobs |k:WHERE|: n = |m:1",
		},
		{
			name: "line comment",
			line: "commit; -- why",
			want: "k:commit|:; |c:-- why",
		},
		{
			name: "string on one line",
			line: "where state = 'queued'",
			want: "k:where|: state = |s:'queued'",
		},
		{
			name: "string left open",
			line: "insert into notes values ('first",
			want: "k:insert|: |k:into|: notes |k:values|: (|s:'first",
			out:  stateSingleQuote,
		},
		{
			name: "string closing",
			in:   stateSingleQuote,
			line: "second');",
			want: "s:second'|:);",
		},
		{
			name: "block comment left open",
			line: "/* the index is",
			want: "c:/* the index is",
			out:  stateBlockComment,
		},
		{
			name: "block comment closing",
			in:   stateBlockComment,
			line: "for the dashboard */ create index",
			want: "c:for the dashboard */|: |k:create|: |k:index",
		},
		{
			name: "quoted identifier",
			line: `select "order" from t`,
			want: `k:select|: |n:"order"|: |k:from|: t`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanSQL(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}
