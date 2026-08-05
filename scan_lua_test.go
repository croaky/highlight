package highlight

import (
	"testing"

	"github.com/croaky/is"
)

func TestScanLua(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "a local and a call",
			line: `local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"`,
			want: `k:local|: lazypath = |n:vim|:.fn.|n:stdpath|:(|s:"data"|:) .. ` +
				`|s:"/lazy/lazy.nvim"`,
		},
		{
			name: "a function and its end",
			line: `local function filetype_autocmd(ft, callback)`,
			want: `k:local|: |k:function|: |n:filetype_autocmd|:(ft, callback)`,
		},
		{
			name: "line comment",
			line: `vim.opt.history = 50 -- keep it short`,
			want: `n:vim|:.opt.history = |m:50|: |c:-- keep it short`,
		},
		{
			// The parens are optional when the only argument is a
			// string, and every plugin spec in a config is written
			// that way.
			name: "a call with no parens",
			line: `require"lazy".setup({})`,
			want: `n:require|s:"lazy"|:.|n:setup|:({})`,
		},
		{
			name: "keywords and operators",
			line: `if not ok and n ~= nil then return end`,
			want: `k:if|: |k:not|: ok |k:and|: n ~= |k:nil|: |k:then|: ` +
				`|k:return|: |k:end`,
		},
		{
			name: "a pattern is a string",
			line: `github_user = github_user:gsub("%s+", "")`,
			want: `:github_user = github_user:|n:gsub|:(|s:"%s+"|:, |s:""|:)`,
		},
		{
			name: "long string left open",
			line: `local sql = [[SELECT id`,
			want: `k:local|: sql = |s:[[SELECT id`,
			out:  stateRawString,
		},
		{
			name: "long string closing",
			in:   stateRawString,
			line: `FROM jobs]])`,
			want: `s:FROM jobs]]|:)`,
		},
		{
			name: "long comment left open",
			line: `--[[ why this exists`,
			want: `c:--[[ why this exists`,
			out:  stateBlockComment,
		},
		{
			name: "long comment closing",
			in:   stateBlockComment,
			line: `and no longer does ]] local x = 1`,
			want: `c:and no longer does ]]|: |k:local|: x = |m:1`,
		},
		{
			name: "an escaped quote does not end the string",
			line: `map("n", "\\", ":Rg<SPACE>")`,
			want: `n:map|:(|s:"n"|:, |s:"\\"|:, |s:":Rg<SPACE>"|:)`,
		},
		{
			name: "hex and float",
			line: `local n = 0xff + 1.5e3 + .5`,
			want: `k:local|: n = |m:0xff|: + |m:1.5e3|: + |m:.5`,
		},
		{
			name: "a goto label is punctuation",
			line: `goto continue ::continue::`,
			want: `k:goto|: continue ::continue::`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanLua(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}
