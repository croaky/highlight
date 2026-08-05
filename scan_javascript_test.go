package highlight

import (
	"strings"
	"testing"

	"github.com/croaky/is"
)

func TestScanJS(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "declaration and numbers",
			line: "const s = 1000, m = 60 * s;",
			want: "k:const|: s = |m:1000|:, m = |m:60|: * s;",
		},
		{
			name: "declaration and call",
			line: "function reltime(t) {",
			want: "k:function|: |n:reltime|:(t) {",
		},
		{
			name: "keyword and string",
			line: "  if (d < 5 * s) return \"<5s ago\";",
			want: ":  |k:if|: (d < |m:5|: * s) |k:return|: |s:\"<5s ago\"|:;",
		},
		{
			name: "line comment",
			line: "el.innerHTML = html; // why",
			want: ":el.innerHTML = html; |c:// why",
		},
		{
			name: "block comment left open",
			line: "const a = 1; /* why",
			want: "k:const|: a = |m:1|:; |c:/* why",
			out:  stateBlockComment,
		},
		{
			name: "block comment closing",
			in:   stateBlockComment,
			line: "still why */ const b = 2;",
			want: "c:still why */|: |k:const|: b = |m:2|:;",
		},
		{
			// A backtick would have closed the Go string this file used
			// to live in, so a template literal is newly expressible,
			// and it is the one string that spans lines.
			name: "template literal left open",
			line: "const q = `hello",
			want: "k:const|: q = |s:`hello",
			out:  stateRawString,
		},
		{
			name: "template literal closing",
			in:   stateRawString,
			line: "world`;",
			want: "s:world`|:;",
		},
		{
			// The slash a value can follow is a regex; the slash an
			// operand can follow is division.
			name: "regex literal",
			line: "if (/^(input|select)$/i.test(t.tagName)) {",
			want: "k:if|: (|s:/^(input|select)$/i|:.|n:test|:(t.tagName)) {",
		},
		{
			name: "division",
			line: "return Math.round(d / s) + \"s ago\";",
			want: "k:return|: |n:Math|:.|n:round|:(d / s) + |s:\"s ago\"|:;",
		},
		{
			name: "an unclosed slash is not a regex",
			line: "const half = total / 2;",
			want: "k:const|: half = total / |m:2|:;",
		},
		{
			name: "escaped quote",
			line: "if (e.key !== \"\\\\\") {",
			want: "k:if|: (e.key !== |s:\"\\\\\"|:) {",
		},
		{
			name: "spread and arrow",
			line: "const open = [...files].some((d) => !d.open);",
			want: "k:const|: open = [...files].|n:some|:((d) => !d.open);",
		},
		{
			name: "await and a global",
			line: "const r = await fetch(el.dataset.live);",
			want: "k:const|: r = |k:await|: |n:fetch|:(el.dataset.live);",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanJS(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}

// TestScanJSByExtension checks the wiring.
func TestScanJSByExtension(t *testing.T) {
	is := is.NewRelaxed(t)

	got := string(Diff("ui/script.js", "@@ -1 +1 @@\n+const a = 1;\n"))
	for _, want := range []string{
		`<span class=k>const</span>`,
		`<span class=m>1</span>`,
	} {
		is.True(strings.Contains(got, want))
	}
}
