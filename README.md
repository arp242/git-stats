git-stats displays aggregate statistics for git repos.

There's a public instance at https://gitstats.arp242.net, with various popular
repos and some things I found interesting, updated semi-regularly. There isn't
really any way to add something to that (yet), partly because this isn't really
finished yet, and partly because importing a repo can be relatively slow so I
need to build something to manage that a bit better. But if you ask nicely I can
add something.

Note there are MANY inaccuracies with all of this:

- A "commit" can mean different things in different projects; I've seen people
  merge small bugfixes with 20 commits, and I've seen people do a huge feature
  in a single commit.

- "Number of commits" doesn't tell the full story; Linus Torvalds is ranked #57
  for Linux and he's not even the top Linus – Linus Walleij is ranked #22 with
  4,338 commits (vs. Torvalds' 2,586).

- Domains may be misleading; for example the top two committers for
  ElasticSearch work for Elastic, but neither use an @elasticsearch.com email
  address.

- Committing code isn't the only way to contribute to a project.

- Some commits aren't code; for example README typo fixes, i18n updates, and
  things like that. For many projects this is a negligible amount of commits,
  but for some it's a large number. It filters some of these out but there is no
  guarantee it filters everything.

- People can use multiple accounts to commit to a project. It tries to merge as
  much as reasonably safe, but this is not 100%. There is also a small chance of
  a false positive in cases where two committers have exactly the same name
  (projects where two notable committers have exactly identical names are
  probably very rare though).

- People's affiliations are not fixed. There are people who committed to Go
  before they were employed at Google, while they were employed at Google, and
  after they were employed at Google. Or any combination of the above.

- Some projects use git "incorrectly" and don't record authorship information in
  the Author header. For example PostgreSQL, Vim, NeoVim, bash, probably more.

But all of that said, I still feel it's useful. Linux is not the average
project, there tends to be a large amount of overlap between the top
contributors and the people doing reviews, support, and the like, and the rest
can be solved by actually look at the data and use your brain before
uncritically accepting any numbers here.

As an aside, GitHub's graph isn't accurate as it only shows commits that can be
linked to a GitHub account and excludes everything else. It's not too uncommon
people commit with an email address not associated with their account, have
their accounts deleted, or things like that. This can be fixed with a .mailmap
file, but many don't bother.

A good example is [mpv]: compare with [git-stats][gmpv], where a major mpv
author deleted their GitHub account, and many of the MPlayer/MPlayer2 authors
don't have a GitHub account. The "top" author on GitHub is actually #11 (and the
margin is huge). The chart also starts in 2010 instead of 2001. All combined it
gives a completely misleading picture.

[mpv]: https://github.com/mpv-player/mpv/graphs/contributors
[gmpv]: https://gitstats.arp242.net/mpv/

Also this codebase isn't super great; I quickly wrote much of this about 4 years
ago to prove a point about something, and I quickly hacked up some more stuff to
look at the contributors for the recent Redis license change. There's tons of
obvious things that can be improved, made less ugly, etc.

Installation
------------
Install from source with:

    % go install zgo.at/git-stats@latest

Or just use `go build` (or install) after a git clone.

Other than a PostgreSQL database, there are no other dependencies.

Usage
-----
You will need a PostgreSQL database, as getting the data out of git is too slow
for many repos, and this drastically reduces storage requirements.

You can also use `git-cache` on local repositories, or it can automatically
clone remote repositories to a cache directory.

The git repo is stored in a cache directory (default: `/tmp/git-stats`). This is
`/tmp` by default as it's ephemeral: you don't need to keep the cache around;
the last commit will be recorded, and on an update it will fetch a shallow clone
from there.

Insert data in the database:

    % git-stats update https://github.com/golang/go

Or a local repo:

    % git-stats update ~/code/my-git-repo

---

There are two interfaces: the CLI and web UI. Start the web UI with:

    % git-stats serve

And most of the interface should be self-explanatory.

You can mix the web and CLI usage: both are backed by the same database.

### CLI usage
You can get stats with e.g.:

    % git-stats author https://github.com/golang/go

Or with the short name:

    % git-stats author go

The `authors` command is the main meat, as this was my primary interest, but
there are also some others:

    % git-stats ls go
    % git-stats activity go

The web interface has a few more features though, basically just out of
laziness.

Alternatives
------------
- https://git-quick-stats.sh

- gitdm – https://lwn.net/Articles/290957/
