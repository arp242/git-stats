package gitstats

import (
	"context"
	"net/url"
	"strings"
	"time"

	"zgo.at/errors"
	"zgo.at/git-stats/db2"
	"zgo.at/zdb"
)

type Repo struct {
	ID            int32      `db:"repo_id,id"`
	Path          string     `db:"path"`
	Name          string     `db:"name"`
	FirstCommit   *Hash      `db:"first_commit"`
	LastCommit    *Hash      `db:"last_commit"`
	FirstCommitAt *time.Time `db:"first_commit_at"`
	LastCommitAt  *time.Time `db:"last_commit_at"`

	Commits int `db:"commits,noinsert"`
}

func (r *Repo) Find(ctx context.Context, name string) error {
	err := r.ByName(ctx, name)
	if err != nil {
		err = r.ByPath(ctx, name)
	}
	return errors.Wrap(err, "Repo.Find")
}

func (r *Repo) ByName(ctx context.Context, name string) error {
	err := zdb.Get(ctx, r, `select * from repos where lower(name) = lower(?)`, name)
	return errors.Wrap(err, "Repo.ByName")
}

func (r *Repo) ByPath(ctx context.Context, path string) error {
	err := zdb.Get(ctx, r, `select * from repos where lower(path) = lower(?)`, path)
	if err != nil {
		// TODO: work around bug where the string values are now all set to ptr
		// to empty string. Need to fix in zdb.
		r.FirstCommit, r.FirstCommitAt, r.LastCommit, r.LastCommitAt = nil, nil, nil, nil
	}
	return errors.Wrap(err, "Repo.ByPath")
}

func (r *Repo) Insert(ctx context.Context) error {
	err := db2.Insert(ctx, r)
	return errors.Wrap(err, "Repo.Insert")
}

func (r *Repo) Update(ctx context.Context) error {
	_, err := db2.Update(ctx, nil, r)
	return errors.Wrap(err, "Repo.Update")
}

func (r Repo) Remote() bool {
	u, err := url.Parse(r.Path)
	if err == nil && u.Scheme != "" {
		return true
	}
	return strings.HasPrefix(r.Path, "git@")
}

type Repos []Repo

func (s *Repos) List(ctx context.Context) error {
	err := zdb.Select(ctx, s, `
		with x as (
			select repo_id, count(*) from commits group by repo_id
		)
		select repos.*, x.count as commits from x
		join repos using (repo_id)
		order by lower(name)
	`)
	return errors.Wrap(err, "Repos.List")
}
