package highlight

import (
	"testing"

	"github.com/croaky/is"
)

func TestScanShell(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "keyword, variable, number",
			line: "if [ $n -gt 3 ]; then",
			want: "k:if|: [ |n:$n|: -gt |m:3|: ]; |k:then",
		},
		{
			name: "comment",
			line: "cp a b # keep",
			want: ":cp a b |c:# keep",
		},
		{
			name: "hash inside a word is not a comment",
			line: "echo ${x#pre}",
			want: ":echo |n:${|:x#pre}",
		},
		{
			name: "quoted strings",
			line: `echo "$HOME" 'raw'`,
			want: `:echo |s:"$HOME"|: |s:'raw'`,
		},
		{
			name: "single quote left open",
			line: "echo 'line one",
			want: ":echo |s:'line one",
			out:  stateSingleQuote,
		},
		{
			name: "single quote closing",
			in:   stateSingleQuote,
			line: "line two' && done",
			want: "s:line two'|: && |k:done",
		},
		{
			name: "double quote left open",
			line: `msg="line one`,
			want: `:msg=|s:"line one`,
			out:  stateDoubleQuote,
		},
		{
			name: "double quote closing",
			in:   stateDoubleQuote,
			line: `line two" # said`,
			want: `s:line two"|: |c:# said`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanShell(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}
