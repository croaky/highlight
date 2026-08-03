#!/usr/bin/env bash
# A comment, and a hash inside a word that is not one.
set -euo pipefail

n=3
echo ${x#pre}
echo "$HOME" 'raw'

msg="a double-quoted string
that spans lines"

echo 'a single-quoted string
that spans lines'

if [ $n -gt 3 ]; then
	for f in *.go; do
		cp "$f" "$f.bak" # keep
	done
else
	exit 1
fi
