package gitstats

import (
	"cmp"
	"context"
	"slices"
	"sort"
	"strings"
	"time"

	"zgo.at/errors"
	"zgo.at/zdb"
	"zgo.at/zstd/ztime"
)

type Author struct {
	ID     int64   `db:"author_id,id"`
	RepoID int32   `db:"repo_id"`
	Names  Strings `db:"names"`
	Emails Strings `db:"emails"`

	Commits int    `db:"-"`
	Added   int    `db:"-"`
	Removed int    `db:"-"`
	First   string `db:"-"`
	Last    string `db:"-"`
}

func (a *Author) ByID(ctx context.Context, repoID int32, authorID int64) error {
	err := zdb.Get(ctx, a, `select * from authors where repo_id=? and author_id=?`, repoID, authorID)
	return errors.Wrap(err, "Author.ByID")
}

// Find author by name or email, or insert if it doesn't exist.
//
// TODO: treat as case-insensitive, now we have:
//
//	Eric Anholt <anholt@freebsd.org>, <anholt@FreeBSD.org>
//	walter harms, Walter Harms <wharms@bfs.de>
//
// And that's not too useful.
func FindOrInsertAuthor(ctx context.Context, c Commit, names, emails map[string]int64) (int64, error) {
	// Don't add the name "unknown"; this completely fucks MariaDB to death
	// because tons of commits have the name set to that with loads of different
	// email addresses.
	//
	// In MySQL tons of email addresses are blank with different names.
	if c.Name == "" || strings.ToLower(c.Name) == "unknown" {
		c.Name = c.Email
	}
	if c.Email == "" {
		c.Email = c.Name
	}

	author, ok := emails[c.Email]
	if !ok {
		author, ok = names[c.Name]
	}

	if !ok {
		fID, err := zdb.InsertID(ctx, "author_id",
			`insert into authors (repo_id, names, emails) values (?, ?, ?)`,
			c.RepoID, Strings{c.Name}, Strings{c.Email})
		if err != nil {
			return 0, errors.Wrapf(err, "FindOrInsertAuthor: Insert: name=%q; email=%q", c.Name, c.Email)
		}
		author = fID
		names[c.Name] = author
		emails[c.Email] = author
		return author, nil
	}

	if _, ok := emails[c.Email]; !ok {
		err := zdb.Exec(ctx,
			`update authors set emails = array_append(emails, ?) where author_id=?`,
			c.Email, author)
		emails[c.Email] = author
		if err != nil {
			return 0, errors.Wrapf(err, "FindOrInsertAuthor: update email: %q", c.Email)
		}
	}
	if _, ok := names[c.Name]; !ok {
		err := zdb.Exec(ctx,
			`update authors set names = array_append(names, ?) where author_id=?`,
			c.Name, author)
		names[c.Name] = author
		if err != nil {
			return 0, errors.Wrapf(err, "FindOrInsertAuthor: update name: %q", c.Name)
		}
	}

	return author, nil
}

type Authors []Author

func (a *Authors) List(ctx context.Context, repoID int32) error {
	err := zdb.Select(ctx, a, `select * from authors where repo_id = ?`, repoID)
	return errors.Wrap(err, "Authors.List")
}

type AuthorStat struct {
	AuthorID    int       `db:"author_id"`
	Commits     int       `db:"commits"`
	Added       int       `db:"added"`
	Removed     int       `db:"removed"`
	CommitPerc  float32   `db:"commit_perc"`
	AddedPerc   float32   `db:"added_perc"`
	RemovedPerc float32   `db:"removed_perc"`
	Names       Strings   `db:"names"`
	Emails      Strings   `db:"emails"`
	First       time.Time `db:"first"`
	Last        time.Time `db:"last"`

	Domains []string `db:"-"`
}

func (s *AuthorStat) ByID(ctx context.Context, repoID int32, authorID int64, rng ztime.Range) error {
	err := zdb.Get(ctx, s, `load:author`, zdb.P{
		"repo_id":   repoID,
		"author_id": authorID,
		"start":     rng.Start,
		"end":       rng.End,
	})
	return errors.Wrap(err, "AuthorStat.ByID")
}

type AuthorStats []AuthorStat

func (s *AuthorStats) List(ctx context.Context, repoID int32, order string, rng ztime.Range) error {
	err := zdb.Select(ctx, s, `load:authors`, zdb.P{
		"repo_id": repoID,
		"start":   rng.Start,
		"end":     rng.End,
	})
	if err != nil {
		return errors.Wrap(err, "AuthorStats.List")
	}

	ss := *s
	if order != "commits" {
		sort.Slice(ss, func(i, j int) bool {
			switch order {
			case "added":
				return ss[i].Added > ss[j].Added
			case "removed":
				return ss[i].Removed > ss[j].Removed
			case "first":
				return ss[j].First.After(ss[i].First)
			case "last":
				return ss[i].Last.After(ss[j].Last)
			default:
				return false
			}
		})
	}

	for i, a := range ss {
		for _, e := range a.Emails {
			j := strings.LastIndexByte(e, '@')
			if j == -1 {
				continue
			}
			if d := e[j:]; !slices.Contains(ss[i].Domains, d) {
				ss[i].Domains = append(ss[i].Domains, d)
			}
		}
	}

	*s = ss
	return nil
}

type Domain struct {
	Domain string
	Count  int
}

func (s AuthorStats) Domains() []Domain {
	domains := make(map[string]int)
	for _, a := range s {
		for _, d := range a.Domains {
			domains[d] += a.Commits
		}
	}
	sortDomains := make([]Domain, 0, len(domains))
	for k, v := range domains {
		sortDomains = append(sortDomains, Domain{k, v})
	}
	slices.SortFunc(sortDomains, func(a, b Domain) int {
		return cmp.Compare(b.Count, a.Count)
	})
	return sortDomains
}
