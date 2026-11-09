package highlight

import "strings"

// patchMeta is the start of a header line in git's extended format.
var patchMeta = []string{
	"index ", "new file mode ", "deleted file mode ", "old mode ",
	"new mode ", "similarity index ", "dissimilarity index ",
	"rename from ", "rename to ", "copy from ", "copy to ",
	"Binary files ",
}

// scanPatch tokenizes one line of a unified diff that is itself the
// file being shown: a .patch or .diff file. A line says what it is by
// its first characters, so no state carries from one line to the next.
//
// The file header lines (diff --git, ---, +++) are keywords, the hunk
// header is a name, and git's extended header lines are comments. An
// added line is a string and a deleted line is a number, which are the
// two classes a reader's eye separates most. Context lines stay plain.
//
// A deleted line that starts with "-- " reads as a file header. Telling
// them apart needs the hunk's line counts, which a one-line scan does
// not have, so that line is colored as a header.
func scanPatch(st state, line string) ([]token, state) {
	var ts tokens
	switch {
	case strings.HasPrefix(line, "@@"):
		ts.add(nameClass, line)
	case strings.HasPrefix(line, "diff --git "),
		strings.HasPrefix(line, "--- "),
		strings.HasPrefix(line, "+++ "):
		ts.add(keywordClass, line)
	case hasAnyPrefix(line, patchMeta):
		ts.add(commentClass, line)
	case strings.HasPrefix(line, "+"):
		ts.add(stringClass, line)
	case strings.HasPrefix(line, "-"):
		ts.add(numberClass, line)
	default:
		ts.add("", line)
	}
	return ts.done(), st
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
