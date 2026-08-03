package highlight

import "testing"

func TestScanHML(t *testing.T) {
	for _, tt := range []struct {
		name string
		line string
		want string
	}{
		{
			// The class and the id are one run: adjacent tokens of the
			// same class merge, since the span would look the same.
			name: "tag with class and id",
			line: "%div.change-side#status",
			want: "k:%div|n:.change-side#status",
		},
		{
			name: "conditional",
			line: "  - if change.title != \"\"",
			want: ":  |k:-|: |k:if|: change.title != |s:\"\"",
		},
		{
			name: "loop",
			line: "- for i, job in jobs",
			want: "k:-|: |k:for|: i, job |k:in|: jobs",
		},
		{
			name: "partial with keyword args",
			line: `= render "check", n: 3`,
			want: `k:=|: |k:render|: |s:"check"|:, n: |m:3`,
		},
		{
			name: "transform is a call",
			line: "= markdown(change.body)",
			want: "k:=|: |n:markdown|:(change.body)",
		},
		{
			name: "output on a tag",
			line: "%td= job.name",
			want: "k:%td=|: job.name",
		},
		{
			name: "comment",
			line: "-# why this row exists",
			want: "c:-# why this row exists",
		},
		{
			name: "filter block header",
			line: ":javascript",
			want: "k::javascript",
		},
		{
			name: "text with interpolation",
			line: "pushed by #{change.author} today",
			want: ":pushed by |n:#{change.author}|: today",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanHML(stateCode, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanHML(%q) = %q, want %q", tt.line, got, tt.want)
			}
			// Nothing in hml spans lines, so the state never moves.
			if out != stateCode {
				t.Errorf("scanHML(%q) ended in state %d, want %d", tt.line, out, stateCode)
			}
		})
	}
}
