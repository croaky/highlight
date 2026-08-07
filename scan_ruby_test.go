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
			name: "any tag terminates the body",
			in:   heredocState("<<~SLACK"),
			line: `  SLACK`,
			want: `s:  SLACK`,
		},
		{
			// The bug the carried tag fixes. A tag no list named
			// used to end at the first line of nothing but
			// capitals, so this body stopped at its own SELECT and
			// the rest of the query was colored as Ruby.
			name: "a bare keyword does not end an unlisted tag",
			in:   heredocState("<<~MIGRATION"),
			line: `  SELECT`,
			want: `s:  SELECT`,
			out:  heredocState("<<~MIGRATION"),
		},
		{
			name: "a quoted heredoc tag",
			line: `  msg = <<-'EOS'`,
			want: `:  msg = |s:<<-'EOS'`,
			out:  heredocState("<<-'EOS'"),
		},
		{
			// The comment ends the line but not the carry, which is
			// the case the pending value has to survive: the body
			// still begins on the next line.
			name: "a comment after a heredoc opener",
			line: `  sql = <<~SQL # why`,
			want: `:  sql = |s:<<~SQL|: |c:# why`,
			out:  heredocState("<<~SQL"),
		},
		{
			// Two on one line. The recursion that used to scan the
			// rest of the line threw its state away, so B was lost
			// and its body came back colored as Ruby.
			name: "two heredocs open on one line",
			line: `a = <<~A; b = <<~B`,
			want: `:a = |s:<<~A|:; b = |s:<<~B`,
			out:  state{kind: kindHeredoc, tag: "A", next: "B"},
		},
		{
			name: "the first body runs while the second waits",
			in:   state{kind: kindHeredoc, tag: "A", next: "B"},
			line: `  first`,
			want: `s:  first`,
			out:  state{kind: kindHeredoc, tag: "A", next: "B"},
		},
		{
			// A's terminator hands over to B rather than to code.
			name: "the first terminator starts the second",
			in:   state{kind: kindHeredoc, tag: "A", next: "B"},
			line: `A`,
			want: `s:A`,
			out:  heredocState("<<~B"),
		},
		{
			name: "the second terminator returns to code",
			in:   heredocState("<<~B"),
			line: `B`,
			want: `s:B`,
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

// TestHeredocQueue covers the second heredoc a line can open and the
// third it cannot.
func TestHeredocQueue(t *testing.T) {
	is := is.NewRelaxed(t)

	var none state
	a := none.queueHeredoc("<<~A")
	is.Eq(a, heredocState("<<~A"))

	ab := a.queueHeredoc("<<~B")
	is.Eq(ab, state{kind: kindHeredoc, tag: "A", next: "B"})

	// The third is dropped rather than displacing either of the first
	// two, since A's body is read before B's and holding any number
	// would cost state its comparability.
	is.Eq(ab.queueHeredoc("<<~C"), ab)

	// A terminator hands over to what is queued, and then to code.
	is.Eq(ab.afterHeredoc(), heredocState("<<~B"))
	is.Eq(heredocState("<<~B").afterHeredoc(), stateCode)
}

// TestHeredocState checks what the opener is reduced to, since the tag
// is now the whole of what a heredoc carry knows.
func TestHeredocState(t *testing.T) {
	is := is.NewRelaxed(t)

	sql := heredocState("<<~SQL")
	is.Eq(sql, state{kind: kindHeredoc, tag: "SQL"})

	// The squiggle, the dash, and the quotes are how the body is
	// indented and interpolated, not part of the word that ends it.
	is.Eq(heredocState("<<-\"SQL\""), sql)
	is.Eq(heredocState("<<'SQL'"), sql)
	is.NotEq(heredocState("<<~HTML"), sql)

	// Every tag is exact, so one must not answer to another's word,
	// and a tag no list would have named is no different.
	is.True(heredocEnds(sql, "  SQL"))
	is.True(!heredocEnds(sql, "  HTML"))
	is.True(!heredocEnds(sql, ""))

	migration := heredocState("<<~MIGRATION")
	is.True(heredocEnds(migration, "  MIGRATION"))
	is.True(!heredocEnds(migration, "  SELECT"))
}
