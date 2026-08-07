package highlight

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/croaky/is"
)

// TestCode checks the whole-file emitter: tokens colored, a carry that
// crosses a line, and the source recoverable from the output.
func TestCode(t *testing.T) {
	is := is.NewRelaxed(t)

	src := "q := `SELECT 1\nFROM jobs`\nfunc f() {}"
	var got strings.Builder
	is.NoErr(Code(&got, "main.go", src))
	out := got.String()

	for _, want := range []string{
		"<span class=s>`SELECT 1</span>",
		"<span class=s>FROM jobs`</span>",
		"<span class=k>func</span>",
	} {
		is.True(strings.Contains(out, want))
	}
	is.Eq(strings.Count(out, "\n"), 2)
	is.True(!strings.Contains(out, "<pre"))
	is.True(!strings.Contains(out, "<code"))
}

// TestCodeByLanguage checks that a bare language name works, which is
// what a markdown fence has instead of a path.
func TestCodeByLanguage(t *testing.T) {
	is := is.NewRelaxed(t)

	var got strings.Builder
	is.NoErr(Code(&got, "haml", "%p= change.title"))
	is.True(strings.Contains(got.String(), "<span class=k>%p=</span>"))
}

// TestCodeEscapes checks that source cannot become markup, in the
// colored path and in the uncolored one. HTML is the source that has to
// survive being read as the language the output is written in: the tag
// is escaped and then colored, so the class span is ours and the
// brackets around it are the reader's text.
func TestCodeEscapes(t *testing.T) {
	is := is.NewRelaxed(t)

	for _, tt := range []struct{ name, want string }{
		{name: "index.html", want: "&lt;<span class=k>script</span>&gt;"},
		{name: "main.go", want: "&lt;script&gt;"},
		{name: "notes.unknownext", want: "&lt;script&gt;"},
	} {
		var got strings.Builder
		is.NoErr(Code(&got, tt.name, "<script>alert(1)</script>"))
		is.True(!strings.Contains(got.String(), "<script>"))
		is.True(strings.Contains(got.String(), tt.want))
	}
}

// BenchmarkCode runs the whole-file emitter over every sample, one
// sub-benchmark each, so a scanner that is slower than its neighbors is
// visible as itself rather than averaged into a total. The samples are
// small and differently sized, which is what SetBytes is for: bytes per
// second compares across them and nanoseconds per operation does not.
//
// Allocations are reported rather than left to -benchmem, since what
// this emitter spends is mostly what it allocates and the flag is a
// thing to remember.
func BenchmarkCode(b *testing.B) {
	names, err := filepath.Glob("testdata/sample.*")
	if err != nil || len(names) == 0 {
		b.Fatal("no samples in testdata")
	}
	for _, name := range names {
		src, err := os.ReadFile(name)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(filepath.Base(name), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(src)))

			for b.Loop() {
				if err := Code(io.Discard, name, string(src)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
