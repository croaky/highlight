package highlight

import (
	"path"
	"strings"
)

// token is a run of source text and the class that colors it. The
// empty class is text that gets no color: operators, punctuation,
// whitespace. A stylesheet paints five classes -- k keywords, n names,
// s strings, m numbers, c comments -- so those are all a scanner emits.
type token struct {
	class string
	text  string
}

// state is what a scanner carries from one line to the next, for the
// tokens that outlive a line: a raw string, a block comment, a fenced
// code block. Its meaning is per-language; only stateCode is shared,
// and it is the zero value so a scan starts mid-nothing.
type state int

const (
	stateCode state = iota
	stateRawString
	stateBlockComment
	stateSingleQuote
	stateDoubleQuote
	stateFence
)

// scanFunc tokenizes one line, given the state the previous line
// ended in, and returns the state this one ended in. Carrying state is
// the whole point: a lexer restarted at every line colors the second
// half of a raw string as code.
type scanFunc func(st state, line string) ([]token, state)

// scannerFor picks a scanner for a filename or a bare language name,
// since the two callers know different things: a diff has a path and a
// markdown fence has a word. A name with an extension is read as a
// path, and everything else is read as the language itself.
//
// A format with no scanner returns nil, which callers render as escaped
// text and no color. That is a diff without code coloring rather than a
// page without a diff: the row keeps its gi/gd class, which is most of
// what the coloring is for.
func scannerFor(name string) scanFunc {
	name = strings.ToLower(name)
	if ext := path.Ext(name); ext != "" {
		name = ext[1:]
	}
	switch name {
	case "go":
		return scanGo
	case "sh", "bash", "zsh", "shell":
		return scanShell
	case "sql":
		return scanSQL
	case "css":
		return scanCSS
	case "js", "mjs", "javascript", "ts", "typescript":
		// TypeScript is JavaScript plus type annotations, which are
		// names and punctuation either way.
		return scanJS
	case "md", "markdown":
		return scanMarkdown
	case "hml", "haml":
		return scanHML
	}
	return nil
}

// tokens accumulates a line's tokens, merging runs of the same class
// so a line of punctuation is one token rather than twenty.
type tokens []token

func (ts *tokens) add(class, text string) {
	if text == "" {
		return
	}
	if n := len(*ts); n > 0 && (*ts)[n-1].class == class {
		(*ts)[n-1].text += text
		return
	}
	*ts = append(*ts, token{class: class, text: text})
}

// The scanners work a byte at a time. A UTF-8 continuation byte is
// never punctuation or a digit, so treating every byte >= 0x80 as part
// of an identifier keeps multibyte text inside one token without
// decoding runes.

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isIdentStart(c byte) bool {
	return c == '_' || c >= 0x80 ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdent(c byte) bool { return isIdentStart(c) || isDigit(c) }

// scanQuoted returns the length of the quoted run starting at s[0],
// which is the quote char, honoring backslash escapes. An unterminated
// quote runs to the end of the line, the way the compiler reads it.
func scanQuoted(s string, quote byte) int {
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case quote:
			return i + 1
		}
	}
	return len(s)
}

// scanNumber returns the length of the numeric literal at the start of
// s. It is deliberately loose: hex digits, base prefixes, underscores,
// exponents and their signs all just continue the number, since the
// only decision downstream is what color to paint it.
func scanNumber(s string) int {
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case isIdent(c) || c == '.':
			i++
		case (c == '+' || c == '-') && i > 0 && isExponent(s[i-1]):
			i++
		default:
			return i
		}
	}
	return i
}

func isExponent(c byte) bool {
	return c == 'e' || c == 'E' || c == 'p' || c == 'P'
}
