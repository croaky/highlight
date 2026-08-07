package highlight

import "strings"

// goKeywords is the spec's list, plus the predeclared constants, which
// read as keywords and are colored like them.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true,
	"continue": true, "default": true, "defer": true, "else": true,
	"fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true,
	"map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true,
	"var": true, "true": true, "false": true, "iota": true, "nil": true,
}

// goNames is the predeclared types and functions. Everything else that
// gets the name color is a call, recognized by the paren after it.
var goNames = map[string]bool{
	"any": true, "bool": true, "byte": true, "comparable": true,
	"complex64": true, "complex128": true, "error": true,
	"float32": true, "float64": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "rune": true,
	"string": true, "uint": true, "uint8": true, "uint16": true,
	"uint32": true, "uint64": true, "uintptr": true,
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true,
	"recover": true,
}

// scanGo tokenizes one line of Go, carrying the two states that outlive
// a line: a raw string and a block comment. Those are the tokens a
// per-line lexer colors as code, which is why this exists.
func scanGo(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateRawString:
			if j := strings.IndexByte(line[i:], '`'); j >= 0 {
				ts.add("s", line[i:i+j+1])
				i += j + 1
				st = stateCode
				continue
			}
			ts.add("s", line[i:])
			return ts, st
		case stateBlockComment:
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
		case c == '/' && i+1 < len(line) && line[i+1] == '/':
			ts.add("c", line[i:])
			return ts, st
		case c == '/' && i+1 < len(line) && line[i+1] == '*':
			ts.add("c", "/*")
			i += 2
			st = stateBlockComment
		case c == '`':
			ts.add("s", "`")
			i++
			st = stateRawString
		case c == '"' || c == '\'':
			// An interpreted string or a rune literal cannot span
			// lines in Go, so an unterminated one ends at the newline.
			n := scanQuoted(line[i:], c)
			ts.add("s", line[i:i+n])
			i += n
		case isDigit(c) || (c == '.' && i+1 < len(line) && isDigit(line[i+1])):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			j := identEnd(line, i+1)
			word := line[i:j]
			switch {
			case goKeywords[word]:
				ts.add("k", word)
			case goNames[word]:
				ts.add("n", word)
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
