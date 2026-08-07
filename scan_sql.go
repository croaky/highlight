package highlight

import "strings"

// sqlKeywords is what ordinary queries and migrations use, matched
// case-insensitively. It is a reading aid, not a dialect definition:
// a word missing here is uncolored, not wrong.
var sqlKeywords = map[string]bool{
	"add": true, "all": true, "alter": true, "and": true, "any": true,
	"as": true, "asc": true, "begin": true, "between": true,
	"bigint": true, "boolean": true, "by": true, "cascade": true,
	"case": true, "cast": true, "check": true, "coalesce": true,
	"column": true, "commit": true, "conflict": true, "constraint": true,
	"count": true, "create": true, "cross": true, "current_timestamp": true,
	"default": true, "delete": true, "desc": true, "distinct": true,
	"do": true, "drop": true, "else": true, "end": true, "exists": true,
	"false": true, "for": true, "foreign": true, "from": true,
	"full": true, "generated": true, "group": true, "having": true,
	"identity": true, "if": true, "in": true, "index": true,
	"inner": true, "insert": true, "int": true, "integer": true,
	"interval": true, "into": true, "is": true, "join": true,
	"jsonb": true, "key": true, "left": true, "like": true,
	"limit": true, "not": true, "now": true, "null": true, "nulls": true,
	"offset": true, "on": true, "or": true, "order": true, "outer": true,
	"primary": true, "references": true, "returning": true,
	"right": true, "rollback": true, "select": true, "set": true,
	"table": true, "text": true, "then": true, "timestamptz": true,
	"true": true, "unique": true, "update": true, "using": true,
	"values": true, "when": true, "where": true, "with": true,
}

// scanSQL tokenizes one line of SQL. A single-quoted string and a
// /* */ comment both span lines, which is the whole reason for the
// carry: a query is an indented block inside a Go raw string, and a
// migration is one statement over many lines.
func scanSQL(st state, line string) ([]token, state) {
	var ts tokens
	for i := 0; i < len(line); {
		switch st {
		case stateSingleQuote:
			if j := strings.IndexByte(line[i:], '\''); j >= 0 {
				ts.add("s", line[i:i+j+1])
				i += j + 1
				st = stateCode
				continue
			}
			ts.add("s", line[i:])
			return ts, st
		case stateBlockComment:
			if j := strings.Index(line[i:], "*/"); j >= 0 {
				ts.add("c", line[i:i+j+2])
				i += j + 2
				st = stateCode
				continue
			}
			ts.add("c", line[i:])
			return ts, st
		}

		c := line[i]
		switch {
		case c == '-' && i+1 < len(line) && line[i+1] == '-':
			ts.add("c", line[i:])
			return ts, st
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
	return ts, st
}
