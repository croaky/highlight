package highlight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			scan := scannerFor(name)
			if scan == nil {
				t.Fatalf("no scanner for %s", name)
			}
			src, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			var st state
			for i, line := range strings.Split(string(src), "\n") {
				ts, next := scan(st, line)
				var got strings.Builder
				for _, tk := range ts {
					got.WriteString(tk.text)
				}
				if got.String() != line {
					t.Fatalf("%s:%d scanned as %q, want %q",
						name, i+1, got.String(), line)
				}
				st = next
			}
			if st != stateCode {
				t.Errorf("scanning %s ended in state %d, want %d",
					name, st, stateCode)
			}
		})
	}
}

// TestScannerForNames checks the two things a caller can hand in: a
// path, whose extension names the language, and the language itself.
func TestScannerForNames(t *testing.T) {
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
		if scannerFor(name) == nil {
			t.Errorf("scannerFor(%q) = nil, want a scanner", name)
		}
	}
	for _, name := range []string{
		"notes.unknownext", "brainfuck", "", "Makefile", "txt", "notes.txt",
	} {
		if scannerFor(name) != nil {
			t.Errorf("scannerFor(%q) returned a scanner, want nil", name)
		}
	}
}
