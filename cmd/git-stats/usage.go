package main

const usage = `Usage: git-stats command [..]
git-stats lists some aggegrate statistics for git repos.
https://github.com/arp242/git-stats

This requires a PostgreSQL database.

Flags:

    -cache      Cache directory for git clones; default: /tmp/git-cache.
    -db         Database connection; default: dbname=git-stats

Commands:

    ls [repo]        List file statistics. 
    activity [repo]  Show activity per month.
    authors [repo]   Show author statistics.
    domains [repo]   Show domain statistics.
    rm [repo]        Remove repo from database.

    serve            Serve HTTP interface.
                       Flags:
                         -listen    Listen address; default: 127.0.0.1:8080
                         -dev       Don't cache results and load templates from
                                    filesystem.

    update [repo]    Update the database from the [repo].

                     If [repo] is a remote URL it will be cloned or updated in
                     the -cache directory.

                     Flags:
                       -no-fetch  Never fetch from remote. Re-use any existing
                                  cache directory for remote URLs.
                       -keep      Keep git cache after the update finished.
                       -name      Short name to use; default is to use
                                  everything after the last /.
`
