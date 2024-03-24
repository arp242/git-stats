#!/usr/bin/env zsh
set -eu

repos=($(psql -XtA git-stats -c "select name from repos"))
integer i
for r in $repos; do
	i+=i
	print -f '\r\x1b[2K\x1b[1m[%d / %d] %s\x1b[0m' $r $i $#repos
	git-stats update $r
done
print
