package highlight

import "testing"

func TestScanJSON(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "a key is a name and a value is a string",
			line: `  "name": "EDS",`,
			want: `:  |n:"name"|:: |s:"EDS"|:,`,
		},
		{
			name: "a key whose value is an object",
			line: `  "compilerOptions": {`,
			want: `:  |n:"compilerOptions"|:: {`,
		},
		{
			name: "words and numbers",
			line: `  "strict": true, "n": 1.5e-3, "x": null`,
			want: `:  |n:"strict"|:: |k:true|:, |n:"n"|:: |m:1.5e-3|:, ` +
				`|n:"x"|:: |k:null`,
		},
		{
			name: "a negative number keeps its sign",
			line: `  "offset": -5`,
			want: `:  |n:"offset"|:: |m:-5`,
		},
		{
			// A colon inside a value is not the value's own.
			name: "a url is a value",
			line: `  "repo": "https://github.com/croaky/highlight",`,
			want: `:  |n:"repo"|:: |s:"https://github.com/croaky/highlight"|:,`,
		},
		{
			name: "an escaped quote does not end the string",
			line: `  "msg": "say \"hi\"",`,
			want: `:  |n:"msg"|:: |s:"say \"hi\""|:,`,
		},
		{
			name: "a string in an array is a value",
			line: `  "files": ["a.ts", "b.ts"],`,
			want: `:  |n:"files"|:: [|s:"a.ts"|:, |s:"b.ts"|:],`,
		},
		{
			// tsconfig.json is the file this is for.
			name: "line comment",
			line: `  "target": "ES2020", // the most minimal diff`,
			want: `:  |n:"target"|:: |s:"ES2020"|:, |c:// the most minimal diff`,
		},
		{
			name: "block comment left open",
			line: `  /* TS 7 prep: these become`,
			want: `:  |c:/* TS 7 prep: these become`,
			out:  stateBlockComment,
		},
		{
			name: "block comment closing",
			in:   stateBlockComment,
			line: `     defaults */ "strict": true`,
			want: `c:     defaults */|: |n:"strict"|:: |k:true`,
		},
		{
			name: "an unterminated string ends at the line",
			line: `  "note": "open`,
			want: `:  |n:"note"|:: |s:"open`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts, out := scanJSON(tt.in, tt.line)
			if got := formatTokens(ts); got != tt.want {
				t.Errorf("scanJSON(%q) = %q, want %q", tt.line, got, tt.want)
			}
			if out != tt.out {
				t.Errorf("scanJSON(%q) ended in state %d, want %d", tt.line, out, tt.out)
			}
		})
	}
}
