package highlight

import "strings"

// cKeywords is C11's reserved words, plus the three the standard
// spells as macros in headers a reader will still read as keywords.
var cKeywords = map[string]bool{
	"auto": true, "break": true, "case": true, "char": true,
	"const": true, "continue": true, "default": true, "do": true,
	"double": true, "else": true, "enum": true, "extern": true,
	"float": true, "for": true, "goto": true, "if": true,
	"inline": true, "int": true, "long": true, "register": true,
	"restrict": true, "return": true, "short": true, "signed": true,
	"sizeof": true, "static": true, "struct": true, "switch": true,
	"typedef": true, "union": true, "unsigned": true, "void": true,
	"volatile": true, "while": true,
	"_Alignas": true, "_Alignof": true, "_Atomic": true, "_Bool": true,
	"_Complex": true, "_Generic": true, "_Imaginary": true,
	"_Noreturn": true, "_Static_assert": true, "_Thread_local": true,
	"bool": true, "false": true, "true": true,
}

// scanC tokenizes one line of C, carrying the one thing that outlives
// a line: a /* */ comment.
//
// A string is not a carry. C only continues one across a line with a
// trailing backslash, which nothing here writes, and the compiler
// reads an unterminated quote as an error either way.
//
// A type name is not colored. The language has a handful of built-in
// ones, which are keywords, and every other type is a typedef the
// header made up, so a list would be right in one file and wrong in
// the next. A call is recognized by the paren after it, as everywhere
// else here.
func scanC(st state, line string) ([]token, state) {
	var ts tokens
	i := 0
	if st == stateCode {
		i = scanCDirective(&ts, line)
	}
	for i < len(line) {
		if st == stateBlockComment {
			if j := strings.Index(line[i:], "*/"); j >= 0 {
				ts.add("c", line[i:i+j+2])
				i += j + 2
				st = stateCode
				continue
			}
			ts.add("c", line[i:])
			return ts, st
		}

		c := line[i]
		switch {
		case strings.HasPrefix(line[i:], "/*"):
			ts.add("c", line[i:i+2])
			i += 2
			st = stateBlockComment
		case strings.HasPrefix(line[i:], "//"):
			ts.add("c", line[i:])
			return ts, st
		case c == '"' || c == '\'':
			n := scanQuoted(line[i:], c)
			ts.add("s", line[i:i+n])
			i += n
		case isDigit(c) || (c == '.' && i+1 < len(line) && isDigit(line[i+1])):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			j := i + 1
			for j < len(line) && isIdent(line[j]) {
				j++
			}
			word := line[i:j]
			switch {
			case cKeywords[word]:
				ts.add("k", word)
			case j < len(line) && line[j] == '(':
				ts.add("n", word)
			default:
				ts.add("", word)
			}
			i = j
		default:
			ts.add("", line[i:i+1])
			i++
		}
	}
	return ts, st
}

// scanCDirective colors a leading preprocessor line and returns where
// the rest of the line starts, or 0 if there is no directive. The
// directive is a keyword, and an include's angle-bracket header is a
// string, since that is what the quoted form of the same line is.
//
// A directive continued with a trailing backslash is not carried: the
// following lines are scanned as code, which is what most of them are.
func scanCDirective(ts *tokens, line string) int {
	i := 0
	for i < len(line) && isSpace(line[i]) {
		i++
	}
	if i == len(line) || line[i] != '#' {
		return 0
	}
	ts.add("", line[:i])

	// The name may be separated from the hash: "#  define" is one
	// directive.
	j := i + 1
	for j < len(line) && isSpace(line[j]) {
		j++
	}
	k := j
	for k < len(line) && isIdent(line[k]) {
		k++
	}
	ts.add("k", line[i:k])

	switch line[j:k] {
	case "include", "include_next":
		p := k
		for p < len(line) && isSpace(line[p]) {
			p++
		}
		if p < len(line) && line[p] == '<' {
			if q := strings.IndexByte(line[p:], '>'); q >= 0 {
				ts.add("", line[k:p])
				ts.add("s", line[p:p+q+1])
				return p + q + 1
			}
		}
	case "error", "warning":
		// The rest is a message, not code. Scanned as code, the
		// apostrophe in "can't" opens a char literal that colors
		// the rest of the line as a string.
		ts.add("", line[k:])
		return len(line)
	}
	return k
}
