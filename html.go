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
	var st state
	for i, line := range strings.Split(src, "\n") {
		if i > 0 {
			buf.WriteByte('\n')
		}
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
		case '@':
			oldSt, newSt = stateCode, stateCode
			buf.WriteString(`<span class=gu>`)
			buf.WriteString(template.HTMLEscapeString(line))
			buf.WriteString(`</span>`)
		case '\\':
			buf.WriteString(`<span class=gu>`)
			buf.WriteString(template.HTMLEscapeString(line))
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
		buf.WriteString(template.HTMLEscapeString(content))
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
			buf.WriteString(template.HTMLEscapeString(t.text))
			continue
		}
		buf.WriteString(`<span class=`)
		buf.WriteString(t.class)
		buf.WriteString(`>`)
		buf.WriteString(template.HTMLEscapeString(t.text))
		buf.WriteString(`</span>`)
	}
}
