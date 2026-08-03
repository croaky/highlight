package sample

import "fmt"

/* A block comment that spans
   lines, which the scanner carries. */

const n = 0x1f + 1.5e-3

// Query is a raw string spanning lines, the case a per-line lexer
// colors as code.
const query = `
	SELECT id
	FROM jobs
	WHERE state = 'queued'
`

func main() {
	var s string = "a\"b"
	if c := s[0]; c == '\n' {
		fmt.Println(query, n)
	}
}
