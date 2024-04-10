#!/usr/bin/env zsh
set -eu

repos=($(psql -XtA git-stats -c "select name from repos"))
integer i
for r in $repos; do
        i+=1
        print -f '\x1b[1m[%d / %d] %s\x1b[0m\n' $i $#repos $r
        ./git-stats -db 'psql+dbname=git-stats sslmode=disable' update $r ||:
done
print
