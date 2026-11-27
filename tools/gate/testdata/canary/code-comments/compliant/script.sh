#!/bin/sh
set -eu
items="a b c"
echo "# not a comment" '# nor this' "${#items}"
printf '%s\n' "$items"
