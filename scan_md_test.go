package highlight

import "testing"

func TestScanMarkdown(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "heading",
			line: "## Why",
			want: "k:## Why",
		},
		{
			name: "inline code and a link",
			line: "see `main.go` and [the plan](todo/planned/014.md)",
			want: ":see |s:`main.go`|: and [|n:the plan|:](todo/planned/014.md)",
		},
		{
			name: "bullet",
			line: "- one",
			want: "k:- |:one",
		},
		{
			name: "ordered item keeps its indent",
			line: "  1. first",
			want: ":  |k:1. |:first",
		},
		{
			name: "quote",
			line: "> quoted",
			want: "c:> quoted",
		},
		{
			name: "fence opening",
			line: "```go",
			want: "c:```go",
			out:  stateFence,
		},
		{
			name: "inside a fence, prose rules do not apply",
			in:   stateFence,
			line: "# not a heading",
			want: ":# not a heading",
			out:  stateFence,
		},
		{
			name: "fence closing",
			in:   stateFence,
			line: "```",
			want: "c:```",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanMarkdown(tt.in, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanMarkdown(%q) = %q, want %q", tt.line, got, tt.want)
			}
			if out != tt.out {
				t.Errorf("scanMarkdown(%q) ended in state %d, want %d", tt.line, out, tt.out)
			}
		})
	}
}
