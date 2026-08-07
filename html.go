package highlight

import (
	"bytes"
	"html/template"
	"io"
	"strings"
)

// Code writes src as HTML-escaped text with each colored run wrapped in
// a class span. name is a filename or a language name; a format with no
// scanner is escaped and left uncolored.
//
// Nothing frames the output: no <pre>, no <code>, no wrapper class. The
// caller already knows what it is putting the code inside, and a markup
// opinion here would be one to work around there.
func Code(w io.Writer, name, src string) error {
	scan := scannerFor(name)

	var buf bytes.Buffer
	buf.Grow(len(src) + len(src)/4)
	var st state
	first := true
	for line := range strings.SplitSeq(src, "\n") {
		if !first {
			buf.WriteByte('\n')
		}
		first = false
		writeContent(&buf, scan, &st, line)
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// Diff renders a unified-diff patch with two layers of coloring: each
// line carries a row-wide diff class (gi inserted, gd deleted, gu hunk
// header, gc context) and, underneath, its content is tokenized by the
// scanner matched to filename.
//
// Style the per-line spans display:block and the row tint extends
// across the full <pre> width, GitHub-style. Marker chars (+/-) inherit
// the row's foreground color; language tokens override it.
//
// A patch is two interleaved programs, so it needs two carry states:
// context plus deletions is the old file, context plus insertions is
// the new one. One state would let an edit inside a string miscolor
// the rest of the hunk. A hunk header resets both, since the file
// jumped.
func Diff(filename, patch string) template.HTML {
	scan := scannerFor(filename)

	var buf bytes.Buffer
	buf.Grow(len(patch) + len(patch)/4)
	buf.WriteString(`<span class=diff>`)
	var oldSt, newSt state
	for line := range strings.SplitSeq(patch, "\n") {
		if line == "" {
			// Trailing empty entry from Split on a patch that ends
			// in "\n"; diff lines always have at least a marker
			// char so empty interior lines don't occur.
			continue
		}
		switch line[0] {
		case '@', '\\':
			if line[0] == '@' {
				// The file jumped, so neither carry means anything.
				oldSt, newSt = stateCode, stateCode
			}
			buf.WriteString(`<span class=gu>`)
			writeEscaped(&buf, line)
			buf.WriteString(`</span>`)
		case '+':
			buf.WriteString(`<span class=gi>+`)
			writeContent(&buf, scan, &newSt, line[1:])
			buf.WriteString(`</span>`)
		case '-':
			buf.WriteString(`<span class=gd>-`)
			writeContent(&buf, scan, &oldSt, line[1:])
			buf.WriteString(`</span>`)
		default:
			// Context line, usually " <content>". It belongs to both
			// files, so both states advance; the new file is the one
			// on screen, so its tokens are the ones written.
			prefix, content := "", line
			if line[0] == ' ' {
				prefix, content = " ", line[1:]
			}
			buf.WriteString(`<span class=gc>`)
			buf.WriteString(prefix)
			same := oldSt == newSt
			if !same && scan != nil {
				_, oldSt = scan(oldSt, content)
			}
			writeContent(&buf, scan, &newSt, content)
			if same {
				oldSt = newSt
			}
			buf.WriteString(`</span>`)
		}
	}
	buf.WriteString(`</span>`)
	return template.HTML(buf.String())
}

// writeContent colors one line's content, advancing st. No scanner
// means no color, not no line.
func writeContent(buf *bytes.Buffer, scan scanFunc, st *state, content string) {
	if scan == nil {
		writeEscaped(buf, content)
		return
	}
	ts, next := scan(*st, content)
	*st = next
	writeTokens(buf, ts)
}

// writeTokens writes tokens as HTML-escaped text, each colored run in a
// class span. The source may be a file a stranger pushed, so every byte
// of it is escaped.
func writeTokens(buf *bytes.Buffer, ts []token) {
	for _, t := range ts {
		if t.class == "" {
			writeEscaped(buf, t.text)
			continue
		}
		buf.WriteString(`<span class=`)
		buf.WriteString(t.class)
		buf.WriteString(`>`)
		writeEscaped(buf, t.text)
		buf.WriteString(`</span>`)
	}
}

// writeEscaped writes s with the bytes that could become markup
// replaced, and the runs between them written as they are. A token with
// nothing to escape is one write and no allocation.
//
// The replacements are the ones html/template writes, byte for byte,
// since the tests spell out expected HTML and the package documents
// this escaping.
func writeEscaped(buf *bytes.Buffer, s string) {
	start := 0
	for i := 0; i < len(s); i++ {
		var repl string
		switch s[i] {
		case 0:
			repl = "\uFFFD"
		case '"':
			repl = "&#34;"
		case '\'':
			repl = "&#39;"
		case '&':
			repl = "&amp;"
		case '<':
			repl = "&lt;"
		case '>':
			repl = "&gt;"
		default:
			continue
		}
		buf.WriteString(s[start:i])
		buf.WriteString(repl)
		start = i + 1
	}
	buf.WriteString(s[start:])
}
