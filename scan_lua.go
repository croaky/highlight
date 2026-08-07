package highlight

import "strings"

// luaKeywords is the reserved words, plus the three literals that read
// as keywords.
var luaKeywords = map[string]bool{
	"and": true, "break": true, "do": true, "else": true,
	"elseif": true, "end": true, "false": true, "for": true,
	"function": true, "goto": true, "if": true, "in": true,
	"local": true, "nil": true, "not": true, "or": true,
	"repeat": true, "return": true, "then": true, "true": true,
	"until": true, "while": true,
}

// luaNames is the standard library and the one global a Neovim config
// is written against. Lua's library is small enough to name, which is
// what makes it worth naming: everything else that gets the name color
// is a call, recognized by the paren after it.
var luaNames = map[string]bool{
	"assert": true, "collectgarbage": true, "coroutine": true,
	"debug": true, "dofile": true, "error": true, "getmetatable": true,
	"io": true, "ipairs": true, "load": true, "loadstring": true,
	"math": true, "next": true, "os": true, "package": true,
	"pairs": true, "pcall": true, "print": true, "rawequal": true,
	"rawget": true, "rawlen": true, "rawset": true, "require": true,
	"select": true, "setmetatable": true, "string": true,
	"table": true, "tonumber": true, "tostring": true, "type": true,
	"unpack": true, "vim": true, "xpcall": true,
}

// scanLua tokenizes one line of Lua, carrying the two things that
// outlive a line: a long string and a long comment, both written in
// double brackets.
//
// Only the level-0 brackets, [[ and ]]. A long string can be opened
// with any number of equals signs -- [==[ closes only at ]==] -- which
// is a depth this carry cannot hold, and which exists for text that
// contains a ]]. Neither the configs nor the articles here have one, and
// the cost of meeting one is code colored as a string until the next
// ]], not a line the reader loses.
//
// A quoted string is not a carry: a newline inside one is an error, and
// a long string is what Lua offers instead.
func scanLua(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateRawString:
			n, closed := ts.drain("s", line[i:], "]]")
			i += n
			if !closed {
				return ts, st
			}
			st = stateCode
			continue
		case stateBlockComment:
			n, closed := ts.drain("c", line[i:], "]]")
			i += n
			if !closed {
				return ts, st
			}
			st = stateCode
			continue
		}

		c := line[i]
		switch {
		case strings.HasPrefix(line[i:], "--[["):
			// Before the line comment below, which it starts with.
			ts.add("c", line[i:i+4])
			i += 4
			st = stateBlockComment
		case strings.HasPrefix(line[i:], "--"):
			ts.add("c", line[i:])
			return ts, st
		case strings.HasPrefix(line[i:], "[["):
			ts.add("s", line[i:i+2])
			i += 2
			st = stateRawString
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
			case luaKeywords[word]:
				ts.add("k", word)
			case luaNames[word]:
				ts.add("n", word)
			case j < len(line) && line[j] == '(':
				ts.add("n", word)
			case j < len(line) && (line[j] == '"' || line[j] == '\''):
				// A call whose only argument is a string may drop
				// the parens: require("x") and require"x" are the
				// same call.
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
