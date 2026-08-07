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

// words is a set of the words given, for the keyword and name lists a
// scanner matches against. A list is the part of a scanner most likely
// to be edited, so it is spelled as the words and nothing else.
func words(w ...string) map[string]bool {
	m := make(map[string]bool, len(w))
	for _, s := range w {
		m[s] = true
	}
	return m
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
	stateTag
	// Last, because Ruby's heredocs take the values above it: one
	// per terminator a carry can name. See heredocTags.
	stateHeredoc
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
	case "json", "jsonc":
		return scanJSON
	case "html", "htm":
		return scanHTML
	case "c", "h":
		// Not cc/cpp/hpp: C++ is a different keyword list, and
		// coloring it with this one names half of it wrong.
		return scanC
	case "lua":
		return scanLua
	case "rb", "ruby":
		return scanRuby
	case "md", "markdown":
		return scanMarkdown
	case "hml", "haml":
		return scanHML
	}
	return nil
}

// tokens accumulates a line's tokens, merging runs of the same class
// so a line of punctuation is one token rather than twenty.
//
// The trailing run is held apart from the finished tokens and becomes
// one when the class changes or the line ends. It used to merge by
// concatenating, which is how this came to account for 92% of the
// allocations rendering a change made: every scanner's default branch
// feeds unclassified bytes one at a time, so a run of punctuation
// allocated once per character and each of those copied the run so far.
//
// A run of one piece keeps the caller's string, which is already a slice
// of the line and costs nothing. Only a run that grew needs a string of
// its own, and buf is reused for the rest of the line.
type tokens struct {
	out   []token
	class string
	text  string // the pending run, while it is a single piece
	buf   []byte // the pending run, once it is more than one
}

// add emits text as class, merging it into the pending run when the
// class has not changed.
func (ts *tokens) add(class, text string) {
	if text == "" {
		return
	}
	if ts.text != "" || len(ts.buf) > 0 {
		if ts.class == class {
			if len(ts.buf) == 0 {
				ts.buf = append(ts.buf, ts.text...)
				ts.text = ""
			}
			ts.buf = append(ts.buf, text...)
			return
		}
		ts.flush()
	}
	ts.class, ts.text = class, text
}

// flush turns the pending run into a token. A run that never grew is
// the caller's own string, so only one that did is copied.
func (ts *tokens) flush() {
	switch {
	case len(ts.buf) > 0:
		ts.out = append(ts.out, token{class: ts.class, text: string(ts.buf)})
		ts.buf = ts.buf[:0]
	case ts.text != "":
		ts.out = append(ts.out, token{class: ts.class, text: ts.text})
		ts.text = ""
	}
}

// done is the line's tokens, the pending run finished. Every scanner
// returns through it, and nothing reads the tokens before it: a run
// still being merged is not one to emit.
func (ts *tokens) done() []token {
	ts.flush()
	return ts.out
}

// drain finishes a carry: it emits s up to and including end and
// reports true, or emits all of s and reports false when end is not on
// this line. The returned length is what to advance by.
//
// Eight scanners end a raw string or a block comment this way, and the
// byte arithmetic is the part that goes wrong. A caller that gets false
// is still inside the carry and returns; one that gets true clears the
// state and keeps scanning the rest of the line.
func (ts *tokens) drain(class, s, end string) (int, bool) {
	if j := strings.Index(s, end); j >= 0 {
		n := j + len(end)
		ts.add(class, s[:n])
		return n, true
	}
	ts.add(class, s)
	return len(s), false
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

func isSpace(c byte) bool { return c == ' ' || c == '\t' }

// identEnd is the index just past the name starting at line[i], which
// the caller has already read as a name's first byte. A language whose
// names take something isIdent does not, a shell flag's hyphen or a
// Ruby method's ?, scans its own.
func identEnd(line string, i int) int {
	for i < len(line) && isIdent(line[i]) {
		i++
	}
	return i
}

// scanDelimited returns the length of the run up to and including
// close, and whether close was found. Escapes count, nesting does not:
// a %w[] with a bracket inside it is rarer than the code this stays
// simple for.
//
// s starts at the first byte of the run, not at an opening delimiter,
// so a caller that has one has to skip it. scanQuoted is that caller.
func scanDelimited(s string, close byte) (int, bool) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case close:
			return i + 1, true
		}
	}
	return len(s), false
}

// scanQuoted returns the length of the quoted run starting at s[0],
// which is the quote char, honoring backslash escapes. An unterminated
// quote runs to the end of the line, the way the compiler reads it, so
// there is nothing for a caller to decide and no bool to return.
func scanQuoted(s string, quote byte) int {
	n, _ := scanDelimited(s[1:], quote)
	return 1 + n
}

// endsOperand reports whether c ends something a slash could divide. A
// name, a digit, or a closing bracket does; an operator, a comma, or an
// open bracket does not, and the slash after one of those opens a
// regex. JavaScript and Ruby both need the rule, and it is the same
// rule in both.
func endsOperand(c byte) bool {
	return c == ')' || c == ']' || c == '}' || c == '$' || isIdent(c)
}

// scanRegex returns the length of the regex literal at the start of s,
// which begins with a slash, or 0 if it does not close on this line. A
// slash inside a character class is a literal slash, which is why the
// class is tracked rather than scanned for the next delimiter.
func scanRegex(s string) int {
	class := false
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '[':
			class = true
		case ']':
			class = false
		case '/':
			if class {
				continue
			}
			i++
			for i < len(s) && isIdent(s[i]) {
				i++
			}
			return i
		}
	}
	return 0
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
