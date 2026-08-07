package highlight

import "strings"

// scanHTML tokenizes one line of HTML: the tag is the keyword, the
// attribute is the name, its value is the string, and text between tags
// is what the page says and takes no color. A document colored
// everywhere is a document a reader cannot skim for markup.
//
// Two carries. A comment spans lines by design. A tag does too, in
// practice: a formatter breaks one tag across as many lines as it has
// attributes, and every hand-written page here has such a tag in its
// head.
//
// An inline script or style is left as text. Coloring it would mean
// running another scanner and holding its carry alongside this one,
// which is a second state this cannot carry -- and the alternative,
// coloring JavaScript as if it were markup, is worse than not coloring
// it. An entity is text for the same reason nothing here invents a
// sixth class for it.
func scanHTML(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateBlockComment:
			n, closed := ts.drain("c", line[i:], "-->")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			continue
		case stateTag:
			n, done := scanTagPart(&ts, line[i:])
			i += n
			if done {
				st = stateCode
			}
			continue
		}

		switch {
		case strings.HasPrefix(line[i:], "<!--"):
			ts.add("c", line[i:i+4])
			i += 4
			st = stateBlockComment
		case line[i] == '<' && i+1 < len(line) && isTagStart(line[i+1]):
			// The bracket and any slash are punctuation; the name
			// after them is the tag. <!doctype and <!DOCTYPE are
			// the same tag to a reader.
			j := i + 1
			if line[j] == '/' || line[j] == '!' {
				j++
			}
			ts.add("", line[i:j])
			k := j
			for k < len(line) && isTagName(line[k]) {
				k++
			}
			ts.add("k", line[j:k])
			i = k
			st = stateTag
		default:
			// Text, up to the next tag. One token, not one per
			// byte, since a paragraph is most of a page.
			j := strings.IndexByte(line[i+1:], '<')
			if j < 0 {
				ts.add("", line[i:])
				return ts.done(), st
			}
			ts.add("", line[i:i+j+1])
			i += j + 1
		}
	}
	return ts.done(), st
}

// scanTagPart reads one piece of the inside of a tag: an attribute, a
// value, or the bracket that closes it. It returns how much it consumed
// and whether the tag ended, so a tag broken across lines resumes here
// on the next one.
func scanTagPart(ts *tokens, s string) (int, bool) {
	switch c := s[0]; {
	case c == '>':
		ts.add("", ">")
		return 1, true
	case c == '/' && strings.HasPrefix(s, "/>"):
		ts.add("", "/>")
		return 2, true
	case c == '"' || c == '\'':
		n := scanQuoted(s, c)
		ts.add("s", s[:n])
		return n, false
	case isTagName(c) && !isDigit(c):
		i := 1
		for i < len(s) && isTagName(s[i]) {
			i++
		}
		ts.add("n", s[:i])
		return i, false
	default:
		// A space, an equals sign, or an unquoted value's first
		// byte. The value is not a string, so it is not colored
		// like one.
		ts.add("", s[:1])
		return 1, false
	}
}

// isTagStart reports whether c can follow the opening bracket of a tag.
// A bracket followed by anything else -- a space, a digit, an operator
// -- is a less-than sign in the page's own text.
func isTagStart(c byte) bool {
	return c == '/' || c == '!' || isIdentStart(c)
}

// isTagName reports whether c continues a tag or attribute name. The
// punctuation is what markup puts in one: a hyphen in a custom element
// or a data attribute, a colon in a namespaced one, a dot and an at
// sign in the framework attributes people write.
func isTagName(c byte) bool {
	switch c {
	case '-', '_', ':', '.', '@':
		return true
	}
	return isIdent(c)
}
