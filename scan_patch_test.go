package highlight

import (
	"strings"
	"testing"

	"github.com/croaky/is"
)

func TestScanPatch(t *testing.T) {
	for _, tt := range []struct {
		name string
		line string
		want string
	}{
		{"a file header", "diff --git a/f.go b/f.go", `k:diff --git a/f.go b/f.go`},
		{"the old file name", "--- a/f.go", `k:--- a/f.go`},
		{"the new file name", "+++ b/f.go", `k:+++ b/f.go`},
		{"a new file has /dev/null", "--- /dev/null", `k:--- /dev/null`},
		{"a hunk header", "@@ -1,3 +1,4 @@ func f() {", `n:@@ -1,3 +1,4 @@ func f() {`},
		{"an index line", "index 846c7d6..f4591b9 100644", `c:index 846c7d6..f4591b9 100644`},
		{"a mode line", "new file mode 100644", `c:new file mode 100644`},
		{"a rename line", "rename from a.go", `c:rename from a.go`},
		{"an added line", "+\tx := 1", "s:+\tx := 1"},
		{"a deleted line", "-\tx := 0", "m:-\tx := 0"},
		{"a context line", " \tx := 2", ":" + " \tx := 2"},
		{"an empty line", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			ts, out := new(tokens).scan(scanPatch, stateCode, tt.line)
			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, stateCode)
		})
	}
}

// A patch file inside a diff is two layers: the outer row class, and
// the scanner of the file's own lines.
func TestDiffOfAPatchFile(t *testing.T) {
	is := is.NewRelaxed(t)
	got := string(Diff("x.patch", "@@ -0,0 +1,3 @@\n+diff --git a/f b/f\n+@@ -1 +1 @@\n+-old\n"))
	for _, want := range []string{
		`<span class=gu>@@ -0,0 +1,3 @@</span>`,
		`<span class=gi>+<span class=k>diff --git a/f b/f</span></span>`,
		`<span class=gi>+<span class=n>@@ -1 +1 @@</span></span>`,
		`<span class=gi>+<span class=m>-old</span></span>`,
	} {
		is.True(strings.Contains(got, want))
	}
}
