package highlight

import "strings"

// rubyKeywords is the reserved words, plus the literals and the two
// visibility words that read as declarations. The rest of what looks
// like a keyword in Ruby is a method -- require, attr_reader, puts, new
// -- and a method is colored as the call it is.
var rubyKeywords = words(
	"alias", "and", "begin", "break", "case", "class", "def",
	"defined?", "do", "else", "elsif", "end", "ensure", "false",
	"for", "if", "in", "module", "next", "nil", "not", "or",
	"private", "protected", "redo", "rescue", "retry", "return",
	"self", "super", "then", "true", "undef", "unless", "until",
	"when", "while", "yield",
)

// scanRuby tokenizes one line of Ruby.
//
// Three carries. A heredoc is the one that matters: the corpus this was
// written against has some fifteen hundred of them, mostly SQL, and a
// scanner that read their bodies as Ruby would color a query's words as
// methods for twenty lines at a time.
//
// A percent literal and a regex are the other two, both of which can
// span lines and rarely do.
//
// Interpolation stays inside its string. #{...} is code, and coloring it
// would mean a scan within a scan; the whole run is one color, the way
// the shell scanner treats $VAR.
func scanRuby(st state, line string) ([]token, state) {
	var ts tokens
	// The last character that mattered, for the slash that is either
	// a regex or a division -- the same ambiguity JavaScript has, and
	// the same rule.
	var prev byte
	for i := 0; i < len(line); {
		switch st {
		case stateCode:
		case stateRawString:
			// A percent literal or a regex left open. Both end at
			// their delimiter, and neither nests across a line.
			if j := strings.IndexAny(line[i:], ")]}>/"); j >= 0 {
				ts.add("s", line[i:i+j+1])
				i += j + 1
				st = stateCode
				continue
			}
			ts.add("s", line[i:])
			return ts.done(), st
		default:
			// A heredoc: the body is the whole line whatever it
			// holds, and so is the terminator, which belongs to the
			// literal it ends.
			ts.add("s", line)
			if heredocEnds(st, line) {
				return ts.done(), stateCode
			}
			return ts.done(), st
		}

		c := line[i]
		switch {
		case c == '#':
			// Not #{...}: interpolation is only inside a string,
			// and a scan is never in code there.
			ts.add("c", line[i:])
			return ts.done(), st
		case c == '"' || c == '\'' || c == '`':
			n := scanQuoted(line[i:], c)
			ts.add("s", line[i:i+n])
			i += n
		case isHeredocStart(line[i:]):
			n := heredocTagLen(line[i:])
			ts.add("s", line[i:i+n])
			st = heredocState(line[i : i+n])
			i += n
			// The rest of the line is code -- `sql = <<~SQL, x` --
			// so the carry starts at the next line, not here.
			ts2, _ := scanRuby(stateCode, line[i:])
			for _, tk := range ts2 {
				ts.add(tk.class, tk.text)
			}
			return ts.done(), st
		case c == '%' && isPercentLiteral(line[i:]):
			n, closed := scanDelimited(line[i+2:], closerFor(line[i+2]))
			ts.add("s", line[i:i+2+n])
			i += 2 + n
			if !closed {
				st = stateRawString
			}
		case c == '/' && !endsOperand(prev):
			if n := scanRegex(line[i:]); n > 0 {
				ts.add("s", line[i:i+n])
				i += n
				break
			}
			ts.add("", line[i:i+1])
			i++
		case strings.HasPrefix(line[i:], "::"):
			// Scope, not a symbol. The constant to its right is
			// colored by the name branch below.
			ts.add("", "::")
			i += 2
		case c == ':' && i+1 < len(line) && isSymbolStart(line[i+1]):
			// A symbol is a name that is also a value, so it takes
			// the name color.
			n := 1 + rubyIdentLen(line[i+1:])
			if q := line[i+1]; q == '"' || q == '\'' {
				// :"two words" is a symbol too.
				n = 1 + scanQuoted(line[i+1:], q)
			}
			ts.add("n", line[i:i+n])
			i += n
		case c == '@' || c == '$':
			// An instance, class, or global variable. The sigil is
			// part of the name.
			j := i + 1
			for j < len(line) && line[j] == '@' {
				j++
			}
			n := j - i + rubyIdentLen(line[j:])
			ts.add("n", line[i:i+n])
			i += n
		case isDigit(c):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			n := rubyIdentLen(line[i:])
			word := line[i : i+n]
			rest := line[i+n:]
			switch {
			case rubyKeywords[word]:
				ts.add("k", word)
			case word[0] >= 'A' && word[0] <= 'Z':
				// A constant: a class, a module, or a constant
				// proper. Ruby tells them apart by the capital
				// and so can a reader.
				ts.add("n", word)
			case strings.HasPrefix(rest, "("):
				ts.add("n", word)
			case strings.HasPrefix(rest, ":") && !strings.HasPrefix(rest, "::"):
				// A hash key written as a label. The colon is
				// punctuation, the way it is in JSON.
				ts.add("n", word)
			default:
				ts.add("", word)
			}
			i += n
		default:
			ts.add("", line[i:i+1])
			i++
		}
		if c != ' ' && c != '\t' {
			prev = c
		}
	}
	return ts.done(), st
}

// rubyIdentLen is the length of the name at the start of s, including a
// trailing ? or ! where that is part of a method's name. empty? and
// save! are names; `x != y` and `cond ? a : b` are not, which is why the
// mark has to touch the name and, for !, not be an inequality. That is
// the whole reason this is not identEnd.
func rubyIdentLen(s string) int {
	i := identEnd(s, 0)
	if i == 0 || i == len(s) {
		return i
	}
	switch s[i] {
	case '?':
		return i + 1
	case '!':
		if i+1 < len(s) && s[i+1] == '=' {
			return i
		}
		return i + 1
	}
	return i
}

// isSymbolStart reports whether c can begin a symbol's name after the
// colon. A quote can: :"two words" is a symbol too.
func isSymbolStart(c byte) bool {
	return c == '"' || c == '\'' || isIdentStart(c)
}

// isHeredocStart reports whether s opens a heredoc. << is also append,
// so the tag is what tells them apart: an uppercase word, optionally
// after ~ or - and optionally quoted. `list << x` is the operator.
func isHeredocStart(s string) bool {
	return heredocTagLen(s) > 0
}

// heredocTagLen is the length of the heredoc opener at the start of s,
// or 0 if there is not one.
func heredocTagLen(s string) int {
	if !strings.HasPrefix(s, "<<") {
		return 0
	}
	i := 2
	if i < len(s) && (s[i] == '~' || s[i] == '-') {
		i++
	}
	var quote byte
	if i < len(s) && (s[i] == '"' || s[i] == '\'') {
		quote = s[i]
		i++
	}
	start := i
	for i < len(s) && (s[i] == '_' || (s[i] >= 'A' && s[i] <= 'Z') || isDigit(s[i])) {
		i++
	}
	if i == start {
		return 0
	}
	if quote != 0 {
		if i >= len(s) || s[i] != quote {
			return 0
		}
		i++
	}
	return i
}

// heredocTags are the terminators a carry can name. state
// stateHeredoc+1+i is "inside a heredoc that ends at heredocTags[i]",
// which is what a body full of capitals needs: a bare FROM on its own
// line does not end a <<~SQL, and formatted SQL is most of what the
// heredocs here hold.
//
// A tag not listed falls back to stateHeredoc, which ends at the first
// line that is nothing but capitals. That is the loose rule, and it is
// right often enough for prose: the bodies whose tags are missing here
// are messages and descriptions, where a line of capitals alone is not
// something anybody writes.
var heredocTags = [...]string{
	"SQL", "GRAPHQL", "HTML", "XML", "HAML", "CSV", "JSON", "YAML",
	"TEXT", "TXT", "DESC", "MSG", "PROMPT", "BODY", "EOS", "EOF",
	"SH", "RUBY", "GO",
}

// heredocState is the carry for a heredoc opened by tag, which is the
// whole opener: <<, any ~ or -, and any quotes.
func heredocState(opener string) state {
	tag := strings.Trim(strings.TrimLeft(opener, "<~-"), `"'`)
	for i, known := range heredocTags {
		if tag == known {
			return stateHeredoc + 1 + state(i)
		}
	}
	return stateHeredoc
}

// heredocEnds reports whether line is the terminator of the heredoc st
// is inside: the tag alone, indented or not.
func heredocEnds(st state, line string) bool {
	s := strings.TrimSpace(line)
	if s == "" {
		return false
	}
	if st > stateHeredoc {
		return s == heredocTags[st-stateHeredoc-1]
	}
	for i := 0; i < len(s); i++ {
		if s[i] != '_' && !(s[i] >= 'A' && s[i] <= 'Z') && !isDigit(s[i]) {
			return false
		}
	}
	return true
}

// isPercentLiteral reports whether s opens one: %w[] and its siblings,
// plus %r for a regex and %q for a string. A bare % is modulo, and %
// followed by a delimiter with no letter is a string nobody writes.
func isPercentLiteral(s string) bool {
	if len(s) < 3 {
		return false
	}
	switch s[1] {
	case 'w', 'W', 'i', 'I', 'q', 'Q', 'r', 's':
	default:
		return false
	}
	return closerFor(s[2]) != 0
}

// closerFor is the delimiter that ends a percent literal opened with
// open, or 0 if open cannot delimit one.
func closerFor(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	case '|', '/', '!', '~':
		return open
	}
	return 0
}
