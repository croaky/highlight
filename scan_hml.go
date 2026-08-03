package highlight

import "strings"

// hmlKeywords is the whole control vocabulary of the language, which
// bans everything else: no assignment, no method calls, no raw output.
var hmlKeywords = map[string]bool{
	"if": true, "elsif": true, "else": true, "for": true, "in": true,
	"render": true, "true": true, "false": true, "nil": true,
}

// scanHML tokenizes one line of hml, the template language whose
// indentation is its structure. Nothing here spans lines: a filter
// block's body is delimited by indentation rather than by a token, so
// there is no carry state to keep.
func scanHML(st state, line string) ([]token, state) {
	var ts tokens
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]
	ts.add("", indent)
	if trimmed == "" {
		return ts, st
	}

	switch {
	case strings.HasPrefix(trimmed, "-#"):
		ts.add("c", trimmed)
	case trimmed[0] == '-' || trimmed[0] == '=':
		ts.add("k", trimmed[:1])
		scanHMLExpr(&ts, trimmed[1:])
	case trimmed[0] == ':':
		// A filter block header: :javascript, :css.
		ts.add("k", trimmed)
	case trimmed[0] == '%' || trimmed[0] == '.' || trimmed[0] == '#':
		scanHMLTag(&ts, trimmed)
	default:
		scanHMLText(&ts, trimmed)
	}
	return ts, st
}

// scanHMLTag colors a tag line: %tag as a keyword, its classes and id
// as names, then whatever follows -- attributes, output, or text.
func scanHMLTag(ts *tokens, s string) {
	i := 0
	if s[0] == '%' {
		j := 1
		for j < len(s) && isIdent(s[j]) {
			j++
		}
		ts.add("k", s[:j])
		i = j
	}
	for i < len(s) && (s[i] == '.' || s[i] == '#') {
		j := i + 1
		for j < len(s) && (isIdent(s[j]) || s[j] == '-') {
			j++
		}
		ts.add("n", s[i:j])
		i = j
	}

	rest := s[i:]
	switch {
	case rest == "":
	case rest[0] == '=':
		ts.add("k", "=")
		scanHMLExpr(ts, rest[1:])
	case rest[0] == '{':
		scanHMLExpr(ts, rest)
	default:
		scanHMLText(ts, rest)
	}
}

// scanHMLExpr colors the expression side of a line: the control words,
// the transforms and helpers a call names, strings, and numbers.
func scanHMLExpr(ts *tokens, s string) {
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '"' || c == '\'':
			n := scanQuoted(s[i:], c)
			ts.add("s", s[i:i+n])
			i += n
		case isDigit(c) && (i == 0 || !isIdent(s[i-1])):
			n := scanNumber(s[i:])
			ts.add("m", s[i:i+n])
			i += n
		case isIdentStart(c):
			j := i + 1
			for j < len(s) && isIdent(s[j]) {
				j++
			}
			word := s[i:j]
			switch {
			case hmlKeywords[word]:
				ts.add("k", word)
			case j < len(s) && s[j] == '(':
				ts.add("n", word)
			default:
				ts.add("", word)
			}
			i = j
		default:
			ts.add("", s[i:i+1])
			i++
		}
	}
}

// scanHMLText colors static text, where the only thing to see is a
// #{field} interpolation.
func scanHMLText(ts *tokens, s string) {
	for {
		i := strings.Index(s, "#{")
		if i < 0 {
			ts.add("", s)
			return
		}
		j := strings.IndexByte(s[i:], '}')
		if j < 0 {
			ts.add("", s)
			return
		}
		ts.add("", s[:i])
		ts.add("n", s[i:i+j+1])
		s = s[i+j+1:]
	}
}
