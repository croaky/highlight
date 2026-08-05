package highlight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/croaky/is"
)

// TestScannersAreLossless runs every scanner over a file in its own
// language and checks that the tokens still spell it. A scanner that
// drops or repeats a byte shows a reader a line nobody wrote, and a
// patch is the one place that is not obvious.
//
// The files also end where they started: a carry left open at the last
// line means the sample has an unterminated string or comment, or the
// scanner lost track of one.
func TestScannersAreLossless(t *testing.T) {
	names, err := filepath.Glob("testdata/sample.*")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("no samples in testdata")
	}
	for _, name := range names {
		t.Run(filepath.Base(name), func(t *testing.T) {
			is := is.New(t)

			scan := scannerFor(name)
			is.NotNil(scan)

			src, err := os.ReadFile(name)
			is.NoErr(err)

			var st state
			for line := range strings.SplitSeq(string(src), "\n") {
				ts, next := scan(st, line)
				var got strings.Builder
				for _, tk := range ts {
					got.WriteString(tk.text)
				}
				is.Eq(got.String(), line)
				st = next
			}
			is.Eq(st, stateCode)
		})
	}
}

// TestScannerForNames checks the two things a caller can hand in: a
// path, whose extension names the language, and the language itself.
func TestScannerForNames(t *testing.T) {
	is := is.NewRelaxed(t)

	for _, name := range []string{
		"main.go", "go", "GO",
		"deploy.sh", "bash", "zsh",
		"schema.sql", "sql",
		"site.css", "css",
		"script.js", "javascript", "ts",
		"tsconfig.json", "jsonc",
		"index.html", "html",
		"init.lua", "lua",
		"person.rb", "ruby",
		"README.md", "markdown",
		"show.hml", "haml",
	} {
		is.NotNil(scannerFor(name))
	}
	for _, name := range []string{
		"notes.unknownext", "brainfuck", "", "Makefile", "txt", "notes.txt",
	} {
		is.Nil(scannerFor(name))
	}
}
