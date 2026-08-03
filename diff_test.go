package highlight

import (
	"fmt"
	"strings"
	"testing"
)

// TestDiffRows checks that every kind of diff line keeps its
// row class and its marker, since those are what color the row and
// what tell an added line from a removed one.
func TestDiffRows(t *testing.T) {
	patch := "@@ -1,3 +1,3 @@\n" +
		" package main\n" +
		"-const a = 1\n" +
		"+const a = 2\n" +
		"\\ No newline at end of file\n"
	got := string(Diff("main.go", patch))

	for _, want := range []string{
		`<span class=gu>@@ -1,3 +1,3 @@</span>`,
		`<span class=gc> `,
		`<span class=gd>-`,
		`<span class=gi>+`,
		`<span class=gu>\ No newline at end of file</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Diff output missing %q\nin: %s", want, got)
		}
	}

	// One span per line and no newlines inside them: the enclosing
	// <pre> ends a line, so a newline here would print a blank row
	// between every pair of diff lines.
	if n := strings.Count(got, "<span class=g"); n != 5 {
		t.Errorf("Diff wrote %d rows, want 5\nin: %s", n, got)
	}
	if strings.Contains(got, "\n") {
		t.Errorf("Diff output contains a newline: %q", got)
	}
}

// TestDiffTokenizes checks the second layer: the scanner
// matched to the filename colors the code inside the row.
func TestDiffTokenizes(t *testing.T) {
	got := string(Diff("main.go", "@@ -1 +1 @@\n+func main() {}\n"))
	if !strings.Contains(got, `<span class=k>func</span>`) {
		t.Errorf("Diff did not tokenize func as a keyword: %s", got)
	}

	// An unknown extension still renders, with no token classes,
	// which is a diff without code coloring rather than a page
	// without a diff.
	got = string(Diff("notes.unknownext", "@@ -1 +1 @@\n+func main() {}\n"))
	if !strings.Contains(got, `<span class=gi>+func main() {}`) {
		t.Errorf("Diff dropped an unlexable line: %s", got)
	}
}

// TestDiffEscapes checks that patch text cannot become
// markup. The content is a file in the repo under review, so it is as
// untrusted as anything else a stranger can push.
func TestDiffEscapes(t *testing.T) {
	got := string(Diff("README.md", "@@ -1 +1 @@\n+<script>alert(1)</script>\n"))
	if strings.Contains(got, "<script>") {
		t.Errorf("Diff passed a script tag through: %s", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Errorf("Diff did not escape a script tag: %s", got)
	}
}

// Highlighting used to be what a change page spent its time on: a
// regexp-driven lexer per line, about 100us of it. The scanners made
// the 502-file change below roughly 400x cheaper, so these benchmarks
// now exist to notice if that is ever given back.

// BenchmarkDiff is one file's patch.
func BenchmarkDiff(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("@@ -1,40 +1,40 @@\n")
	for i := range 40 {
		fmt.Fprintf(&sb, "-\tif err := doThing(%d, \"x\"); err != nil {\n", i)
		fmt.Fprintf(&sb, "+\tif err := doThing(%d, \"y\"); err != nil {\n", i)
	}
	patch := sb.String()

	for b.Loop() {
		Diff("main.go", patch)
	}
}

// BenchmarkDiffChange is a whole change's worth of files, shaped
// like the 502-file one that made the page take half a minute: many
// files, each a small patch. This is the number to watch, since a
// change page renders every file before it answers.
func BenchmarkDiffChange(b *testing.B) {
	var patches []string
	for f := range 502 {
		var sb strings.Builder
		sb.WriteString("@@ -1,7 +1,7 @@\n")
		for i := range 7 {
			fmt.Fprintf(&sb, "-\tfoo := bar(%d, \"baz\") // comment\n", i)
		}
		fmt.Fprintf(&sb, "+\tfoo := bar(%d, \"qux\")\n", f)
		patches = append(patches, sb.String())
	}

	for b.Loop() {
		for _, p := range patches {
			Diff("main.go", p)
		}
	}
}
