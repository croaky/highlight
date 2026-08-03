package highlight

import "strings"

// scanCSS tokenizes one line of CSS.
//
// What makes a word a property here is the colon after it, not the
// brace it is inside. A scanner that counted braces would be right on a
// whole file and wrong on a patch: a hunk starts wherever the diff
// says, so half of them open inside a rule with no `{` in sight, and
// every declaration in that hunk would read as a selector. A colon
// with something other than a name after it is a declaration's; a colon
// a name runs into is a pseudo-class. That holds line by line.
//
// The comment is the one thing that carries, and the one that matters:
// a stylesheet worth reading is more comment than rule.
func scanCSS(st state, line string) ([]token, state) {
	var ts tokens
	// Whether what follows, up to the next semicolon, is a value. The
	// same words appear on both sides of a colon -- a value's are
	// values, and coloring them as the selectors they are elsewhere
	// paints most of a stylesheet one color.
	value := false
	for i := 0; i < len(line); {
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
		case c == '/' && i+1 < len(line) && line[i+1] == '*':
			ts.add("c", "/*")
			i += 2
			st = stateBlockComment
		case c == '"' || c == '\'':
			n := scanQuoted(line[i:], c)
			ts.add("s", line[i:i+n])
			i += n
		case c == '@':
			n := 1 + scanCSSIdent(line[i+1:])
			ts.add("k", line[i:i+n])
			i += n
		case c == '#':
			// A color, or the id of a selector. Hex is the ambiguity:
			// #fff is a color and #f00d is not an id, so the digits and
			// their count decide.
			if n := scanHexColor(line[i:]); n > 0 {
				ts.add("m", line[i:i+n])
				i += n
				break
			}
			n := 1 + scanCSSIdent(line[i+1:])
			ts.add("n", line[i:i+n])
			i += n
		case isPseudo(line[i:]):
			// A pseudo-class or pseudo-element, which belongs to the
			// selector it hangs off. A declaration's colon has a space
			// after it and falls through to punctuation.
			j := i + 1
			if line[j] == ':' {
				j++
			}
			j += scanCSSIdent(line[j:])
			ts.add("n", line[i:j])
			i = j
		case (c == '.' || c == '-') && i+1 < len(line) && isDigit(line[i+1]):
			n := 1 + scanNumber(line[i+1:])
			ts.add("m", line[i:i+n])
			i += n
		case isDigit(c):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case c == '.' && i+1 < len(line) && isCSSIdentStart(line[i+1]):
			// The dot is part of the name: .diff-file is one thing to
			// read, not a name behind an operator.
			n := 1 + scanCSSIdent(line[i+1:])
			ts.add("n", line[i:i+n])
			i += n
		case isCSSIdentStart(c):
			j := i + scanCSSIdent(line[i:])
			word := line[i:j]
			switch {
			case j < len(line) && line[j] == ':' && !isPseudo(line[j:]):
				ts.add("k", word)
				value = true
			case value && strings.HasPrefix(word, "--"):
				// A custom property being read. It is a name here and a
				// property where it is declared, which is what it is.
				ts.add("n", word)
			case value:
				ts.add("", word)
			default:
				ts.add("n", word)
			}
			i = j
		default:
			if c == ';' || c == '{' || c == '}' {
				value = false
			}
			ts.add("", line[i:i+1])
			i++
		}
	}
	return ts, st
}

// isPseudo reports whether s begins with the colon of a pseudo-class or
// pseudo-element rather than the colon of a declaration. One function
// for both readings, because they are the same question asked from
// either side of the colon, and two spellings of it disagreed: ::after
// is a pseudo-element, so the word in front of it is a selector and not
// a property.
func isPseudo(s string) bool {
	return len(s) > 1 && s[0] == ':' && (s[1] == ':' || isCSSIdentStart(s[1]))
}

// isCSSIdentStart reports whether c can begin a CSS name. The hyphen
// can: a vendor prefix and a custom property both lead with one, and
// -webkit-user-select is one name rather than a subtraction.
func isCSSIdentStart(c byte) bool { return c == '-' || isIdentStart(c) }

func isCSSIdent(c byte) bool { return isCSSIdentStart(c) || isDigit(c) }

// scanCSSIdent returns the length of the name at the start of s.
func scanCSSIdent(s string) int {
	i := 0
	for i < len(s) && isCSSIdent(s[i]) {
		i++
	}
	return i
}

// scanHexColor returns the length of the hex color at the start of s,
// which begins with #, or 0 if what follows is not one. CSS allows
// three, four, six, and eight digits; anything else with a # in front
// of it is an id.
func scanHexColor(s string) int {
	i := 1
	for i < len(s) && isHexDigit(s[i]) {
		i++
	}
	switch n := i - 1; n {
	case 3, 4, 6, 8:
		if i < len(s) && isCSSIdent(s[i]) {
			return 0
		}
		return i
	}
	return 0
}

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
