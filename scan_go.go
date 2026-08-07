package highlight

// goKeywords is the spec's list, plus the predeclared constants, which
// read as keywords and are colored like them.
var goKeywords = words(
	"break", "case", "chan", "const", "continue", "default",
	"defer", "else", "fallthrough", "for", "func", "go", "goto",
	"if", "import", "interface", "map", "package", "range",
	"return", "select", "struct", "switch", "type", "var", "true",
	"false", "iota", "nil",
)

// goNames is the predeclared types and functions. Everything else that
// gets the name color is a call, recognized by the paren after it.
var goNames = words(
	"any", "bool", "byte", "comparable", "complex64", "complex128",
	"error", "float32", "float64", "int", "int8", "int16", "int32",
	"int64", "rune", "string", "uint", "uint8", "uint16", "uint32",
	"uint64", "uintptr", "append", "cap", "clear", "close",
	"complex", "copy", "delete", "imag", "len", "make", "max",
	"min", "new", "panic", "print", "println", "real", "recover",
)

// scanGo tokenizes one line of Go, carrying the two states that outlive
// a line: a raw string and a block comment. Those are the tokens a
// per-line lexer colors as code, which is why this exists.
func scanGo(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateRawString:
			n, closed := ts.drain("s", line[i:], "`")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			continue
		case stateBlockComment:
			n, closed := ts.drain("c", line[i:], "*/")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			continue
		}

		c := line[i]
		switch {
		case c == '/' && i+1 < len(line) && line[i+1] == '/':
			ts.add("c", line[i:])
			return ts.done(), st
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
	return ts.done(), st
}
