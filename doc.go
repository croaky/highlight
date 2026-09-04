// Package highlight colors source code for server-rendered HTML.
//
// It is a hand-written, line-oriented scanner per language over a
// shared token type, for a curated set of formats: Go, sh, SQL, CSS,
// JavaScript, C, Ruby, JSON, Lua, HTML, Markdown, and hml. A format it
// has no scanner for is escaped and rendered uncolored, so callers can
// mix.
//
// # Scope
//
// Five classes come out of a scanner -- k keywords, n names, s
// strings, m numbers, c comments -- and nothing else. The curation is
// the product: this is fast and small because it is deliberately
// incomplete and covers only a handful of languages. It is not a lexer
// framework, not a plugin registry, and not correctness-complete for
// any language. A word a keyword list misses is uncolored, not wrong.
//
// # Two layers
//
// Tokens are the lower layer: a scanFunc reads one line, given the
// state the line before it ended in, and returns the state this one
// ended in. Carrying state is the whole point -- a lexer restarted per
// line colors the second half of a raw string as code.
//
// Emitters are the upper layer, and HTML is one consumer rather than
// the interface. Code writes a whole file; Diff writes a unified-diff
// patch inside one span of class diff, where each row also carries a
// row class (gi inserted, gd deleted, gu hunk header, gc context).
//
// The two do not share a signature, and the difference is the
// distinction rather than an oversight. Code takes an io.Writer because
// a whole file is something a caller streams and frames: it writes its
// own <pre> and <code> around the call, with no string in between. Diff
// returns template.HTML because a patch is a fragment that becomes a
// value: a page renders one per file into a row a template reads, where
// a writer would mean a buffer for each.
//
// So a third emitter takes the shape of what it emits rather than one
// this package picked. A whole file takes a writer; a fragment a
// template will hold returns one. Nothing needs both shapes for the same
// thing, and offering both on each emitter would double the surface to
// save a caller a line.
//
// # Escaping
//
// Both emitters HTML-escape every byte they write. The source may be a
// file a stranger pushed, so nothing in it can become markup.
//
// # Styling
//
// The ten class names are the whole contract with a stylesheet, and
// there is no shipped palette: colors are the consuming site's.
package highlight
