package highlight

import (
	"strings"
	"testing"
)

func TestScanCSS(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "selector, properties, and lengths",
			line: ".page-text { margin: 0.7em 0.5em 1em; font-family: monospace; }",
			want: "n:.page-text|: { |k:margin|:: |m:0.7em|: |m:0.5em|: " +
				"|m:1em|:; |k:font-family|:: monospace; }",
		},
		{
			// What a hunk starting inside a rule looks like. The
			// property is a property because of the colon after it, not
			// because of a brace on some line the patch left out.
			name: "declaration with no brace in sight",
			line: "\ttop: calc(var(--head-h) - 1px);",
			want: ":\t|k:top|:: calc(var(|n:--head-h|:) - |m:1px|:);",
		},
		{
			name: "comment left open",
			line: "/* The sticky bar's height: its own box, and the",
			want: "c:/* The sticky bar's height: its own box, and the",
			out:  stateBlockComment,
		},
		{
			name: "inside a comment",
			in:   stateBlockComment,
			line: "   height the browser works out is not a number",
			want: "c:   height the browser works out is not a number",
			out:  stateBlockComment,
		},
		{
			name: "comment closing",
			in:   stateBlockComment,
			line: "   is left. */ :root { --side-w: 320px; }",
			want: "c:   is left. */|: |n::root|: { |k:--side-w|:: |m:320px|:; }",
		},
		{
			// A colon a word runs into is a pseudo-class, not the colon
			// of a declaration, so the word before it is a selector.
			name: "pseudo-class and pseudo-element",
			line: ".diff-file > summary:hover::after { background: #f6f8fa; }",
			want: "n:.diff-file|: > |n:summary:hover::after|: { " +
				"|k:background|:: |m:#f6f8fa|:; }",
		},
		{
			// The word before a :: is a selector too, which the two
			// readings of a colon have to agree on.
			name: "vendor pseudo-element",
			line: ".diff-file > summary::-webkit-details-marker { display: none; }",
			want: "n:.diff-file|: > |n:summary::-webkit-details-marker|: { " +
				"|k:display|:: none; }",
		},
		{
			name: "an id selector is not a color",
			line: "#checkout-cmd { color: #fff; }",
			want: "n:#checkout-cmd|: { |k:color|:: |m:#fff|:; }",
		},
		{
			name: "at-rule and media feature",
			line: "@media (min-width: 800px) {",
			want: "k:@media|: (|k:min-width|:: |m:800px|:) {",
		},
		{
			name: "element and attribute selectors",
			line: "button, input[type=submit] {",
			want: "n:button|:, |n:input|:[|n:type|:=|n:submit|:] {",
		},
		{
			name: "vendor prefix and a bare value",
			line: ".row-id a { -webkit-user-select: all; }",
			want: "n:.row-id|: |n:a|: { |k:-webkit-user-select|:: all; }",
		},
		{
			name: "negative and fractional lengths",
			line: "\tmargin: -1px 0 .5em;",
			want: ":\t|k:margin|:: |m:-1px|: |m:0|: |m:.5em|:;",
		},
		{
			name: "quoted value",
			line: "\tcontent: \"x\";",
			want: ":\t|k:content|:: |s:\"x\"|:;",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanCSS(tt.in, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanCSS(%q) = %q, want %q", tt.line, got, tt.want)
			}
			if out != tt.out {
				t.Errorf("scanCSS(%q) ended in state %d, want %d", tt.line, out, tt.out)
			}
		})
	}
}

// TestScanCSSByExtension checks the wiring: a .css patch is tokenized
// rather than merely tinted, which is the whole point of having the
// scanner.
func TestScanCSSByExtension(t *testing.T) {
	got := string(Diff("ui/style.css", "@@ -1 +1 @@\n+.a { color: #fff; }\n"))
	for _, want := range []string{
		`<span class=n>.a</span>`,
		`<span class=k>color</span>`,
		`<span class=m>#fff</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("a css patch is missing %q\nin: %s", want, got)
		}
	}
}
