package highlight

import (
	"strings"
	"testing"
)

// TestCode checks the whole-file emitter: tokens colored, a carry that
// crosses a line, and the source recoverable from the output.
func TestCode(t *testing.T) {
	src := "q := `SELECT 1\nFROM jobs`\nfunc f() {}"
	var got strings.Builder
	if err := Code(&got, "main.go", src); err != nil {
		t.Fatal(err)
	}
	out := got.String()

	for _, want := range []string{
		"<span class=s>`SELECT 1</span>",
		"<span class=s>FROM jobs`</span>",
		"<span class=k>func</span>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Code output missing %q\nin: %s", want, out)
		}
	}
	if n := strings.Count(out, "\n"); n != 2 {
		t.Errorf("Code wrote %d newlines, want 2\nin: %s", n, out)
	}
	if strings.Contains(out, "<pre") || strings.Contains(out, "<code") {
		t.Errorf("Code framed its output: %s", out)
	}
}

// TestCodeByLanguage checks that a bare language name works, which is
// what a markdown fence has instead of a path.
func TestCodeByLanguage(t *testing.T) {
	var got strings.Builder
	if err := Code(&got, "haml", "%p= change.title"); err != nil {
		t.Fatal(err)
	}
	if want := "<span class=k>%p=</span>"; !strings.Contains(got.String(), want) {
		t.Errorf("Code(haml) missing %q\nin: %s", want, got.String())
	}
}

// TestCodeEscapes checks that source cannot become markup, in the
// colored path and in the uncolored one.
func TestCodeEscapes(t *testing.T) {
	for _, name := range []string{"index.html", "main.go"} {
		var got strings.Builder
		if err := Code(&got, name, "<script>alert(1)</script>"); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(got.String(), "<script>") {
			t.Errorf("Code(%q) passed a script tag through: %s", name, got.String())
		}
		if !strings.Contains(got.String(), "&lt;script&gt;") {
			t.Errorf("Code(%q) did not escape a script tag: %s", name, got.String())
		}
	}
}
