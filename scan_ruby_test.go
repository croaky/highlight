package highlight

import (
	"testing"

	"github.com/croaky/is"
)

func TestScanRuby(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   state
		line string
		want string
		out  state
	}{
		{
			name: "a class and a constant",
			line: `class EnqueueIngest < Jobs::Base`,
			want: `k:class|: |n:EnqueueIngest|: < |n:Jobs|:::|n:Base`,
		},
		{
			name: "a def and an instance variable",
			line: `  def initialize(db) @db = db end`,
			want: `:  |k:def|: |n:initialize|:(db) |n:@db|: = db |k:end`,
		},
		{
			name: "symbols and labels",
			line: `  Jobs::Insert.new.call(queue: "harmonic", name: :ingest)`,
			// A call with no parens is not colored, which is the
			// same rule the JavaScript scanner uses: .new here is
			// indistinguishable from an attribute.
			want: `:  |n:Jobs|:::|n:Insert|:.new.|n:call|:(|n:queue|:: ` +
				`|s:"harmonic"|:, |n:name|:: |n::ingest|:)`,
		},
		{
			name: "a comment",
			line: `  # why this exists`,
			want: `:  |c:# why this exists`,
		},
		{
			// Interpolation is code, but coloring it would be a
			// scan inside a scan, so the string keeps its color --
			// and the # inside one is not a comment.
			name: "interpolation stays in the string",
			line: `  puts "count: #{rows.size} of #{total}"`,
			want: `:  puts |s:"count: #{rows.size} of #{total}"`,
		},
		{
			name: "a method with a question mark",
			line: `  return true if rows.empty?`,
			want: `:  |k:return|: |k:true|: |k:if|: rows.empty?`,
		},
		{
			name: "an inequality is not a bang method",
			line: `  save! if a != b`,
			want: `:  save! |k:if|: a != b`,
		},
		{
			name: "a ternary is not a question method",
			line: `  n = ok ? 1 : 2`,
			want: `:  n = ok ? |m:1|: : |m:2`,
		},
		{
			name: "a word list",
			line: `  ATTRS = %w[name city state].freeze`,
			want: `:  |n:ATTRS|: = |s:%w[name city state]|:.freeze`,
		},
		{
			name: "percent literals take any delimiter",
			line: `  re = %r{\A/api/} and q = %q(one) + %Q<two>`,
			want: `:  re = |s:%r{\A/api/}|: |k:and|: q = |s:%q(one)|: + |s:%Q<two>`,
		},
		{
			name: "a percent literal left open",
			line: `  ATTRS = %w[name`,
			want: `:  |n:ATTRS|: = |s:%w[name`,
			out:  stateRawString,
		},
		{
			name: "a percent literal closing",
			in:   stateRawString,
			line: `    city].freeze`,
			want: `s:    city]|:.freeze`,
		},
		{
			name: "a backtick runs a command",
			line: "  sha = `git rev-parse HEAD`.strip",
			want: ":  sha = |s:`git rev-parse HEAD`|:.strip",
		},
		{
			name: "a regex and a division",
			line: `  if name =~ /\A[a-z]+\z/ then n / 2 end`,
			want: `:  |k:if|: name =~ |s:/\A[a-z]+\z/|: |k:then|: n / |m:2|: |k:end`,
		},
		{
			name: "append is not a heredoc",
			line: `  rows << row`,
			want: `:  rows << row`,
		},
		{
			name: "a heredoc opens and the line goes on",
			line: `  db.exec(<<~SQL, id)`,
			want: `:  db.|n:exec|:(|s:<<~SQL|:, id)`,
			out:  heredocState("<<~SQL"),
		},
		{
			// The reason the tag is remembered: a bare SQL keyword
			// on its own line is not the end of the body.
			name: "a sql keyword is not the terminator",
			in:   heredocState("<<~SQL"),
			line: `    FROM`,
			want: `s:    FROM`,
			out:  heredocState("<<~SQL"),
		},
		{
			name: "the terminator ends it",
			in:   heredocState("<<~SQL"),
			line: `  SQL`,
			want: `s:  SQL`,
		},
		{
			// A tag nothing here names falls back to the loose
			// rule: the first line that is nothing but capitals.
			name: "an unnamed tag ends at any capitals",
			in:   heredocState("<<~SLACK"),
			line: `  SLACK`,
			want: `s:  SLACK`,
		},
		{
			name: "a quoted heredoc tag",
			line: `  msg = <<-'EOS'`,
			want: `:  msg = |s:<<-'EOS'`,
			out:  heredocState("<<-'EOS'"),
		},
		{
			name: "numbers",
			line: `  n = 0xff + 1_000 + 1.5e3`,
			want: `:  n = |m:0xff|: + |m:1_000|: + |m:1.5e3`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			is := is.NewRelaxed(t)

			ts, out := scanRuby(tt.in, tt.line)

			is.Eq(formatTokens(ts), tt.want)
			is.Eq(out, tt.out)
		})
	}
}

// TestHeredocState checks the two answers a tag can get: a state of its
// own for the tags worth naming, and the shared loose one otherwise.
func TestHeredocState(t *testing.T) {
	is := is.NewRelaxed(t)

	sql := heredocState("<<~SQL")
	is.True(sql > stateHeredoc)
	is.Eq(heredocState("<<-\"SQL\""), sql)
	is.NotEq(heredocState("<<~HTML"), sql)
	is.Eq(heredocState("<<~SLACK"), stateHeredoc)

	// The loose rule is what an unnamed tag gets, and a named one
	// must not answer to another tag's word.
	is.True(!heredocEnds(sql, "  HTML"))
	is.True(heredocEnds(stateHeredoc, "  ANYTHING"))
	is.True(!heredocEnds(stateHeredoc, ""))
}
