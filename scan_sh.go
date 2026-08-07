package highlight

import "strings"

// shKeywords is the shell's own words. Builtins are not here: which
// names a shell resolves internally is a long list that changes per
// shell, and a reader does not read a script for that.
var shKeywords = words(
	"if", "then", "elif", "else", "fi", "for", "while", "until",
	"do", "done", "case", "esac", "in", "function", "select",
	"time", "return", "break", "continue", "local", "export",
	"readonly", "set", "shift", "trap", "exit",
)

// scanShell tokenizes one line of sh. Both quote kinds span lines, so
// both are carry states. $VAR inside a double-quoted string stays part
// of the string: the whole run is one color anyway.
func scanShell(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateSingleQuote:
			if j := strings.IndexByte(line[i:], '\''); j >= 0 {
				ts.add("s", line[i:i+j+1])
				i += j + 1
				st = stateCode
				continue
			}
			ts.add("s", line[i:])
			return ts, st
		case stateDoubleQuote:
			if n, closed := scanDelimited(line[i:], '"'); closed {
				ts.add("s", line[i:i+n])
				i += n
				st = stateCode
				continue
			}
			ts.add("s", line[i:])
			return ts, st
		}

		c := line[i]
		switch {
		case c == '#' && (i == 0 || isSpace(line[i-1])):
			// A # elsewhere is part of a word: ${x#y}, a fragment in
			// a URL.
			ts.add("c", line[i:])
			return ts, st
		case c == '\'':
			ts.add("s", "'")
			i++
			st = stateSingleQuote
		case c == '"':
			ts.add("s", `"`)
			i++
			st = stateDoubleQuote
		case c == '$':
			j := i + 1
			for j < len(line) && isIdent(line[j]) {
				j++
			}
			if j == i+1 && j < len(line) && (line[j] == '{' || line[j] == '(') {
				j++
			}
			ts.add("n", line[i:j])
			i = j
		case isDigit(c) && (i == 0 || !isIdent(line[i-1])):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			j := i + 1
			for j < len(line) && (isIdent(line[j]) || line[j] == '-') {
				j++
			}
			if word := line[i:j]; shKeywords[word] {
				ts.add("k", word)
			} else {
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
