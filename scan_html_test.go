package highlight

import "testing"

func TestScanHTML(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "a tag, an attribute, and text",
			line: `<a href="/about">About</a>`,
			want: `:<|k:a|: |n:href|:=|s:"/about"|:>About</|k:a|:>`,
		},
		{
			name: "the doctype is a tag",
			line: `<!doctype html>`,
			want: `:<!|k:doctype|: |n:html|:>`,
		},
		{
			name: "a void element closes itself",
			line: `<meta charset="UTF-8" />`,
			want: `:<|k:meta|: |n:charset|:=|s:"UTF-8"|: />`,
		},
		{
			// A formatter breaks a tag with several attributes
			// across a line each, which every page here has.
			name: "a tag left open",
			line: `    <meta`,
			want: `:    <|k:meta`,
			out:  stateTag,
		},
		{
			name: "an attribute on its own line",
			in:   stateTag,
			line: `      name="viewport"`,
			want: `:      |n:name|:=|s:"viewport"`,
			out:  stateTag,
		},
		{
			name: "the line that closes the tag",
			in:   stateTag,
			line: `    />text`,
			want: `:    />text`,
		},
		{
			name: "a data attribute keeps its hyphens",
			line: `<input ah-submit-on-type ah-debounce-on-type="300" />`,
			want: `:<|k:input|: |n:ah-submit-on-type|: ` +
				`|n:ah-debounce-on-type|:=|s:"300"|: />`,
		},
		{
			name: "an unquoted value is not a string",
			line: `<details class=faq open>`,
			want: `:<|k:details|: |n:class|:=|n:faq|: |n:open|:>`,
		},
		{
			name: "comment left open",
			line: `<!-- why this exists`,
			want: `c:<!-- why this exists`,
			out:  stateBlockComment,
		},
		{
			name: "comment closing",
			in:   stateBlockComment,
			line: `and no longer does --><p>next</p>`,
			want: `c:and no longer does -->|:<|k:p|:>next</|k:p|:>`,
		},
		{
			// A bracket that starts no tag is the page's own text,
			// which is most of what a page about HTML contains.
			name: "a less-than sign is text",
			line: `if a < b and b > c then`,
			want: `:if a < b and b > c then`,
		},
		{
			name: "a single-quoted value",
			line: `<div id='main'>`,
			want: `:<|k:div|: |n:id|:=|s:'main'|:>`,
		},
		{
			name: "an entity is text",
			line: `<p>a &amp; b</p>`,
			want: `:<|k:p|:>a &amp; b</|k:p|:>`,
		},
		{
			// A template's own syntax is not markup and gets no
			// color, which keeps it visible against the attribute
			// around it.
			name: "a template action inside a value",
			line: `<link rel="stylesheet" href="{{.CSSPath}}" />`,
			want: `:<|k:link|: |n:rel|:=|s:"stylesheet"|: ` +
				`|n:href|:=|s:"{{.CSSPath}}"|: />`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanHTML(tt.in, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanHTML(%q) = %q, want %q", tt.line, got, tt.want)
			}
			if out != tt.out {
				t.Errorf("scanHTML(%q) ended in state %d, want %d", tt.line, out, tt.out)
			}
		})
	}
}
