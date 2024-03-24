package gitstats

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"zgo.at/errors"
	"zgo.at/zdb"
	"zgo.at/zstd/ztime"
)

type Commit struct {
	RepoID     int32        `db:"repo_id"`
	Hash       Hash         `db:"hash,id"`
	Date       time.Time    `db:"date"`
	AuthorID   int          `db:"author_id"`
	Exclude    sql.NullBool `db:"exclude"`
	Files      Strings      `db:"files"`
	Added      Ints         `db:"added"`
	Removed    Ints         `db:"removed"`
	AddedSum   int          `db:"added_sum"`
	RemovedSum int          `db:"removed_sum"`
	Subject    string       `db:"subject"`

	Email    string       `db:"-"`
	Name     string       `db:"-"`
	UpdFiles []CommitFile `db:"files"`
}
type CommitFile struct {
	Path, Added, Removed string
	Exclude              sql.NullBool
}

var (
	reTypo  = regexp.MustCompile(`\btypo\b`)
	reExt   = regexp.MustCompile(`\.(?:md|markdown|rst|txt|adoc|asciidoc|json|yml|yaml|po)$`)
	reFile  = regexp.MustCompile(`/(?:LICENSE|README|COPYING|INSTALL|NEWS|POTFILES|LINGUAS|Change[lL]og?).*`)
	reFile2 = regexp.MustCompile(`/(?:configure(?:\.ac|\.in)?|config\.(?:guess|sub)|ltconfig|ltmain\.sh|Makefile\.(?:in|am))$`)
)

func (c *CommitFile) ShouldExclude() bool {
	if c.Exclude.Valid {
		return c.Exclude.Bool
	}
	if !reFile.MatchString(c.Path) && !reExt.MatchString(c.Path) && !reFile2.MatchString(c.Path) {
		c.Exclude.Valid, c.Exclude.Bool = true, false
		return false
	}
	c.Exclude.Valid, c.Exclude.Bool = true, true
	return true
}

// Also: toml 782628a7 (gofmt -s)
func (c *Commit) ShouldExclude() bool {
	if c.Exclude.Valid {
		return c.Exclude.Bool
	}

	if reTypo.MatchString(c.Subject) {
		c.Exclude.Valid, c.Exclude.Bool = true, true
		return true
	}
	for _, f := range c.UpdFiles {
		if !f.ShouldExclude() { // Only if *all* files match exclude.
			c.Exclude.Valid, c.Exclude.Bool = true, false
			return false
		}
	}
	c.Exclude.Valid, c.Exclude.Bool = true, true
	return true
}

type Commits []Commit

func (c *Commits) ByAuthor(ctx context.Context, repoID int32, authorID int64, rng ztime.Range) error {
	err := zdb.Select(ctx, c, `load:commits`, zdb.P{
		"repo_id":   repoID,
		"author_id": authorID,
		"start":     rng.Start,
		"end":       rng.End,
	})
	return errors.Wrap(err, "Commits.ByAuthor")
}

type CommitStat struct {
	Date    time.Time `db:"date" json:"date"`
	Commits int       `db:"commits" json:"commits"`
	Added   int       `db:"added" json:"added"`
	Removed int       `db:"removed" json:"removed"`
}

type CommitStats []CommitStat

func (s *CommitStats) List(ctx context.Context, repoID int32, rng ztime.Range, authorID int64, groupMonth bool) error {
	err := zdb.Select(ctx, s, `load:activity`, zdb.P{
		"repo_id":   repoID,
		"author_id": authorID,
		"start":     rng.Start,
		"end":       rng.End,
		"by_day":    !groupMonth,
	})
	return errors.Wrap(err, "CommitStat.List")
}

func CountCommits(ctx context.Context, repoID int32, rng ztime.Range) (int, error) {
	var n int
	err := zdb.Get(ctx, &n, `
		select count(*) from commits where repo_id=:repo_id and exclude=0
		{{:start and date >= :start}}
		{{:end   and date <= :end}}
	`, zdb.P{
		"repo_id": repoID,
		"start":   rng.Start,
		"end":     rng.End,
	})
	return n, errors.Wrap(err, "CountCommits")
}
