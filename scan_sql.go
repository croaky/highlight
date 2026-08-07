package highlight

import "strings"

// sqlKeywords is what ordinary queries and migrations use, matched
// case-insensitively. It is a reading aid, not a dialect definition:
// a word missing here is uncolored, not wrong.
var sqlKeywords = words(
	"add", "all", "alter", "and", "any", "as", "asc", "begin",
	"between", "bigint", "boolean", "by", "cascade", "case", "cast",
	"check", "coalesce", "column", "commit", "conflict",
	"constraint", "count", "create", "cross", "current_timestamp",
	"default", "delete", "desc", "distinct", "do", "drop", "else",
	"end", "exists", "false", "for", "foreign", "from", "full",
	"generated", "group", "having", "identity", "if", "in", "index",
	"inner", "insert", "int", "integer", "interval", "into", "is",
	"join", "jsonb", "key", "left", "like", "limit", "not", "now",
	"null", "nulls", "offset", "on", "or", "order", "outer",
	"primary", "references", "returning", "right", "rollback",
	"select", "set", "table", "text", "then", "timestamptz", "true",
	"unique", "update", "using", "values", "when", "where", "with",
)

// scanSQL tokenizes one line of SQL. A single-quoted string and a
// /* */ comment both span lines, which is the whole reason for the
// carry: a query is an indented block inside a Go raw string, and a
// migration is one statement over many lines.
func scanSQL(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateSingleQuote:
			n, closed := ts.drain("s", line[i:], "'")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			continue
		case stateBlockComment:
			n, closed := ts.drain("c", line[i:], "*/")
			i += n
			if !closed {
				return ts.done(), st
			}
			st = stateCode
			continue
		}

		c := line[i]
		switch {
		case c == '-' && i+1 < len(line) && line[i+1] == '-':
			ts.add("c", line[i:])
			return ts.done(), st
		case c == '/' && i+1 < len(line) && line[i+1] == '*':
			ts.add("c", "/*")
			i += 2
			st = stateBlockComment
		case c == '\'':
			ts.add("s", "'")
			i++
			st = stateSingleQuote
		case c == '"':
			// A quoted identifier, not a string: it names a column.
			n := scanQuoted(line[i:], '"')
			ts.add("n", line[i:i+n])
			i += n
		case isDigit(c):
			n := scanNumber(line[i:])
			ts.add("m", line[i:i+n])
			i += n
		case isIdentStart(c):
			j := identEnd(line, i+1)
			if word := line[i:j]; sqlKeywords[strings.ToLower(word)] {
				ts.add("k", word)
			} else {
				ts.add("", word)
			}
			i = j
		default:
			ts.add("", line[i:i+1])
			i++
		}
	}
	return ts.done(), st
}
