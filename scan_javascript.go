package highlight

// jsKeywords is the reserved words, plus the literals that read as
// keywords. The contextual ones -- get, set, from, static -- are left
// out: they are ordinary method names often enough that coloring them
// would be wrong more than right.
var jsKeywords = words(
	"async", "await", "break", "case", "catch", "class", "const",
	"continue", "debugger", "default", "delete", "do", "else",
	"export", "extends", "false", "finally", "for", "function",
	"if", "import", "in", "instanceof", "let", "new", "null", "of",
	"return", "super", "switch", "this", "throw", "true", "try",
	"typeof", "undefined", "var", "void", "while", "with", "yield",
)

// jsNames is the globals a page reaches for. They are the closest
// thing JavaScript has to Go's predeclared identifiers; everything else
// that gets the name color is a call, recognized by the paren after it.
var jsNames = words(
	"AbortController", "Array", "Boolean", "DOMParser", "Date",
	"Error", "EventSource", "FormData", "JSON", "Map", "Math",
	"Number", "Object", "Promise", "RegExp", "Set", "String", "URL",
	"URLSearchParams", "clearInterval", "clearTimeout", "console",
	"document", "fetch", "history", "localStorage", "location",
	"navigator", "queueMicrotask", "requestAnimationFrame",
	"sessionStorage", "setInterval", "setTimeout",
	"structuredClone", "window",
)

// scanJS tokenizes one line of JavaScript, carrying the two states that
// outlive a line: a template literal and a block comment. A quoted
// string is not one of them, since a newline inside one is an error.
//
// A slash is the language's ambiguity: it opens a regex where a value
// belongs and divides where an operand just ended. What precedes it
// decides, so the scan remembers the last character that was not
// whitespace -- an identifier, a number, or a closing bracket means
// division, and anything else means a regex. That is the same rule a
// parser uses, minus the parser.
func scanJS(st state, line string) ([]token, state) {
	var ts tokens
	// The last character that mattered, for the slash above. Zero at
	// the start of a line, which reads as "a value belongs here": a
	// line beginning with a regex is a line beginning with a value.
	var prev byte
	for i := 0; i < len(line); {
		switch st {
		case stateRawString:
			n, closed := ts.drain("s", line[i:], "`")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			// A template literal is a value, so a slash after it
			// divides.
			prev = '`'
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
		case c == ' ' || c == '\t':
			ts.add("", line[i:i+1])
			i++
			continue
		case c == '/' && i+1 < len(line) && line[i+1] == '/':
			ts.add("c", line[i:])
			return ts.done(), st
		case c == '/' && i+1 < len(line) && line[i+1] == '*':
			ts.add("c", "/*")
			i += 2
			st = stateBlockComment
		case c == '/' && !endsOperand(prev):
			// A regex, if it closes on this line. One that does not is
			// a division whose right side is still being typed.
			if n := scanRegex(line[i:]); n > 0 {
				ts.add("s", line[i:i+n])
				i += n
				break
			}
			ts.add("", line[i:i+1])
			i++
		case c == '`':
			ts.add("s", "`")
			i++
			st = stateRawString
		case c == '"' || c == '\'':
			n := scanQuoted(line[i:], c)
			ts.add("s", line[i:i+n])
			i += n
		case isDigit(c) || (c == '.' && i+1 < len(line) && isDigit(line[i+1])):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isJSIdentStart(c):
			j := i + 1
			for j < len(line) && isJSIdent(line[j]) {
				j++
			}
			word := line[i:j]
			switch {
			case jsKeywords[word]:
				ts.add("k", word)
			case jsNames[word]:
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
		prev = c
	}
	return ts.done(), st
}

// isJSIdentStart reports whether c can begin a name. The dollar sign
// can, and is a name of its own in plenty of code.
func isJSIdentStart(c byte) bool { return c == '$' || isIdentStart(c) }

func isJSIdent(c byte) bool { return isJSIdentStart(c) || isDigit(c) }
