package highlight

import "strings"

// scanMarkdown tokenizes one line of Markdown, which is the worst fit
// of the five and worth saying so: Markdown is block structure, not
// tokens. This does the little that survives a line at a time --
// headings, fences, list markers, quotes, inline code, link text --
// and a line scanner will never see a setext underline coming.
//
// The fence is the one thing here that needs the carry, and it is the
// one that matters: a diff inside a fenced block is not prose.
func scanMarkdown(st state, line string) ([]token, state) {
	var ts tokens
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
		ts.add("", indent)
		ts.add("c", trimmed)
		if st == stateFence {
			return ts.done(), stateCode
		}
		return ts.done(), stateFence
	}
	if st == stateFence {
		// Code, in a language we were not told. Uncolored is honest.
		ts.add("", line)
		return ts.done(), st
	}

	switch {
	case strings.HasPrefix(trimmed, "#"):
		ts.add("", indent)
		ts.add("k", trimmed)
		return ts.done(), st
	case strings.HasPrefix(trimmed, ">"):
		ts.add("", indent)
		ts.add("c", trimmed)
		return ts.done(), st
	}

	ts.add("", indent)
	rest := trimmed
	if n := markdownMarker(rest); n > 0 {
		ts.add("k", rest[:n])
		rest = rest[n:]
	}
	scanMarkdownInline(&ts, rest)
	return ts.done(), st
}

// markdownMarker returns the length of a leading list marker: a bullet
// or an ordered number, each followed by a space.
func markdownMarker(s string) int {
	if len(s) > 1 && (s[0] == '-' || s[0] == '*' || s[0] == '+') && s[1] == ' ' {
		return 2
	}
	i := 0
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	if i > 0 && i+1 < len(s) && (s[i] == '.' || s[i] == ')') && s[i+1] == ' ' {
		return i + 2
	}
	return 0
}

// scanMarkdownInline colors the two inline spans worth having: `code`
// and the text of a [link](url). Emphasis is left alone, since its
// markers are also ordinary punctuation.
func scanMarkdownInline(ts *tokens, s string) {
	for i := 0; i < len(s); {
		switch s[i] {
		case '`':
			if j := strings.IndexByte(s[i+1:], '`'); j >= 0 {
				ts.add("s", s[i:i+j+2])
				i += j + 2
				continue
			}
			ts.add("", s[i:])
			return
		case '[':
			if j := strings.IndexByte(s[i:], ']'); j > 0 {
				ts.add("", "[")
				ts.add("n", s[i+1:i+j])
				ts.add("", "]")
				i += j + 1
				continue
			}
			ts.add("", s[i:i+1])
			i++
		default:
			ts.add("", s[i:i+1])
			i++
		}
	}
}
