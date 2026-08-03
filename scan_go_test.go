package highlight

import (
	"strings"
	"testing"
)

// TestScanGo prices one line at a time: a keyword, a name, a string, a
// number, a comment, and -- the case the old per-line lexer got wrong
// -- a token opened on one line and closed on the next.
func TestScanGo(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "keyword and call",
			line: "func main() {}",
			want: "k:func|: |n:main|:() {}",
		},
		{
			name: "predeclared type",
			line: "var n int = 42",
			want: "k:var|: n |n:int|: = |m:42",
		},
		{
			name: "string with an escaped quote",
			line: `s := "a\"b" + c`,
			want: `:s := |s:"a\"b"|: + c`,
		},
		{
			name: "rune literal",
			line: "if c == '\\n' {",
			want: "k:if|: c == |s:'\\n'|: {",
		},
		{
			name: "line comment",
			line: "x++ // why",
			want: ":x++ |c:// why",
		},
		{
			name: "block comment closed on the same line",
			line: "a /* mid */ b",
			want: ":a |c:/* mid */|: b",
		},
		{
			name: "block comment left open",
			line: "a /* mid",
			want: ":a |c:/* mid",
			out:  stateBlockComment,
		},
		{
			name: "inside a block comment",
			in:   stateBlockComment,
			line: "still a comment",
			want: "c:still a comment",
			out:  stateBlockComment,
		},
		{
			name: "block comment closing",
			in:   stateBlockComment,
			line: "done */ func f() {",
			want: "c:done */|: |k:func|: |n:f|:() {",
		},
		{
			name: "raw string left open",
			line: "q := `SELECT 1",
			want: ":q := |s:`SELECT 1",
			out:  stateRawString,
		},
		{
			name: "inside a raw string",
			in:   stateRawString,
			line: "\tFROM jobs WHERE id = 1",
			want: "s:\tFROM jobs WHERE id = 1",
			out:  stateRawString,
		},
		{
			name: "raw string closing",
			in:   stateRawString,
			line: "`, id)",
			want: "s:`|:, id)",
		},
		{
			name: "hex and float",
			line: "n := 0x1f + 1.5e-3",
			want: ":n := |m:0x1f|: + |m:1.5e-3",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanGo(tt.in, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanGo(%q) = %q, want %q", tt.line, got, tt.want)
			}
			if out != tt.out {
				t.Errorf("scanGo(%q) ended in state %d, want %d", tt.line, out, tt.out)
			}
		})
	}
}

// formatTokens spells a token list as class:text runs joined by |, so a
// table test reads as one line per line of source.
func formatTokens(ts []token) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, t.class+":"+t.text)
	}
	return strings.Join(parts, "|")
}

// TestDiffCarriesStatePerSide checks the two-state rule: an
// edit inside a raw string leaves the old and new files in different
// states, and neither may leak into the other.
func TestDiffCarriesStatePerSide(t *testing.T) {
	patch := "@@ -1,3 +1,3 @@\n" +
		" q := `SELECT 1\n" +
		"-FROM old`\n" +
		"+FROM new`\n" +
		" func f() {}\n"
	got := string(Diff("main.go", patch))

	// Both edited lines are inside the string on their own side.
	for _, want := range []string{
		`<span class=gd>-<span class=s>FROM old`,
		`<span class=gi>+<span class=s>FROM new`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Diff output missing %q\nin: %s", want, got)
		}
	}
	// The string closed on both sides, so the context line after it is
	// code again.
	if !strings.Contains(got, `<span class=gc> <span class=k>func</span>`) {
		t.Errorf("Diff kept scanning past a closed raw string: %s", got)
	}
}
