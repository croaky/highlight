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
// colored path and in the uncolored one. HTML is the source that has to
// survive being read as the language the output is written in: the tag
// is escaped and then colored, so the class span is ours and the
// brackets around it are the reader's text.
func TestCodeEscapes(t *testing.T) {
	for _, tt := range []struct{ name, want string }{
		{name: "index.html", want: "&lt;<span class=k>script</span>&gt;"},
		{name: "main.go", want: "&lt;script&gt;"},
		{name: "notes.unknownext", want: "&lt;script&gt;"},
	} {
		var got strings.Builder
		if err := Code(&got, tt.name, "<script>alert(1)</script>"); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(got.String(), "<script>") {
			t.Errorf("Code(%q) passed a script tag through: %s", tt.name, got.String())
		}
		if !strings.Contains(got.String(), tt.want) {
			t.Errorf("Code(%q) missing %q\nin: %s", tt.name, tt.want, got.String())
		}
	}
}
