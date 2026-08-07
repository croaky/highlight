package highlight

import "strings"

// jsonWords is every bare word JSON has. Anything else unquoted is
// not JSON, so it stays uncolored rather than being guessed at.
var jsonWords = map[string]bool{
	"true": true, "false": true, "null": true,
}

// scanJSON tokenizes one line of JSON. A key is a name and a value is
// a string, which is the one distinction worth making in a language
// with no keywords: a reader scans the left column for the field they
// want.
//
// Comments are read, though JSON has none. The files people call JSON
// and hand to a compiler -- tsconfig.json is the example -- carry both
// kinds, and coloring one as punctuation and its text as fields would
// be worse than reading a dialect nothing here has to parse. A string
// cannot span lines in either dialect, so the block comment is the
// only carry.
func scanJSON(st state, line string) ([]token, state) {
	var ts tokens
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
		case c == '/' && i+1 < len(line) && line[i+1] == '/':
			ts.add("c", line[i:])
			return ts, st
		case c == '/' && i+1 < len(line) && line[i+1] == '*':
			st = stateBlockComment
			ts.add("c", line[i:i+2])
			i += 2
		case c == '"':
			n := scanQuoted(line[i:], '"')
			ts.add(jsonStringClass(line[i+n:]), line[i:i+n])
			i += n
		case isDigit(c) || (c == '-' && i+1 < len(line) && isDigit(line[i+1])):
			// The minus belongs to the number it signs. Nothing in
			// JSON subtracts, so there is no other reading.
			n := 1 + scanNumber(line[i+1:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			j := i + 1
			for j < len(line) && isIdent(line[j]) {
				j++
			}
			if word := line[i:j]; jsonWords[word] {
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

// jsonStringClass is n for a key and s for a value, told apart by what
// follows the closing quote: a colon makes the string a field name.
// rest is the line from there, so a key whose colon is on the next line
// reads as a value -- nobody writes that, and the alternative is a
// carry state for a guess.
func jsonStringClass(rest string) string {
	for i := 0; i < len(rest); i++ {
		switch {
		case isSpace(rest[i]):
		case rest[i] == ':':
			return "n"
		default:
			return "s"
		}
	}
	return "s"
}
