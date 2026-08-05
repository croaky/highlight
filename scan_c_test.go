package highlight

import (
	"testing"

	"github.com/croaky/is"
)

func TestScanC(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "an angle-bracket include",
			line: `#include <stdlib.h>`,
			want: `k:#include|: |s:<stdlib.h>`,
		},
		{
			name: "a quoted include",
			line: `#include "tree_sitter/parser.h"`,
			want: `k:#include|: |s:"tree_sitter/parser.h"`,
		},
		{
			name: "a define is a directive and code",
			line: `#define MAX_DEPTH 64`,
			want: `k:#define|: MAX_DEPTH |m:64`,
		},
		{
			// The hash is the directive; the name may sit apart
			// from it.
			name: "an indented directive",
			line: `  #  ifdef __APPLE__`,
			want: `:  |k:#  ifdef|: __APPLE__`,
		},
		{
			// The message is prose, so an apostrophe in it is not
			// a char literal running to the newline.
			name: "an error message is not code",
			line: `#error can't build without a compiler`,
			want: `k:#error|: can't build without a compiler`,
		},
		{
			name: "a declaration and a call",
			line: `static void advance(TSLexer *lexer) { lexer->advance(lexer, false); }`,
			want: `k:static|: |k:void|: |n:advance|:(TSLexer *lexer) { lexer->` +
				`|n:advance|:(lexer, |k:false|:); }`,
		},
		{
			name: "line comment",
			line: `s->len = 1; // start over at the top level`,
			want: `:s->len = |m:1|:; |c:// start over at the top level`,
		},
		{
			name: "block comment left open",
			line: `x = 1; /* why this exists`,
			want: `:x = |m:1|:; |c:/* why this exists`,
			out:  stateBlockComment,
		},
		{
			name: "block comment closing",
			in:   stateBlockComment,
			line: `and no longer does */ return 0;`,
			want: `c:and no longer does */|: |k:return|: |m:0|:;`,
		},
		{
			// A directive inside a comment is comment, so the
			// carry is read before the line's first byte.
			name: "a directive inside a block comment",
			in:   stateBlockComment,
			line: `#include <stdio.h>`,
			want: `c:#include <stdio.h>`,
			out:  stateBlockComment,
		},
		{
			name: "a char literal and its escapes",
			line: `if (c == '\n' || c == '\'') return true;`,
			want: `k:if|: (c == |s:'\n'|: || c == |s:'\''|:) |k:return|: |k:true|:;`,
		},
		{
			name: "hex, suffix, and float",
			line: `uint16_t n = 0xffUL + 1.5e3 + .5f;`,
			want: `:uint16_t n = |m:0xffUL|: + |m:1.5e3|: + |m:.5f|:;`,
		},
		{
			name: "a sizeof is a keyword, not a call",
			line: `size = sizeof(uint8_t) + s->len;`,
			want: `:size = |k:sizeof|:(uint8_t) + s->len;`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanC(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}
