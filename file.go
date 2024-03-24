package gitstats

import (
	"context"
	"strconv"

	"zgo.at/errors"
	"zgo.at/zdb"
)

type File struct {
	ID      int64  `db:"file_id,id"`
	RepoID  int    `db:"repo_id"`
	Path    string `db:"path"`
	Exclude bool   `db:"exclude"`
}

type Files []File

func (f *Files) List(ctx context.Context, repoID int32) error {
	err := zdb.Select(ctx, f, `select * from files where repo_id = ?`, repoID)
	return errors.Wrap(err, "Files.List")
}

// Map returns a path → file_id map.
func (f *Files) Map() map[string]int64 {
	m := make(map[string]int64, len(*f))
	for _, ff := range *f {
		m[ff.Path] = ff.ID
	}
	return m
}

type FileStat []struct {
	ID         int    `db:"file_id"`
	Path       string `db:"path"`
	NumCommits string `db:"num_commits"`
}

// TODO:
// positional: "/foo/bar%"
// -depth   group stuff by max depth
// Add authors info
func (s *FileStat) List(ctx context.Context, repoID int32) error {
	err := zdb.Select(ctx, s, `load:ls`, zdb.P{"repo_id": repoID})
	return errors.Wrap(err, "FileStat.List")
}

func FindOrInsertFile(ctx context.Context, c Commit, files map[string]int64) ([]int64, Ints, Ints, error) {
	var (
		commitFiles = make([]int64, 0, len(c.UpdFiles))
		added       = make(Ints, 0, len(c.UpdFiles))
		removed     = make(Ints, 0, len(c.UpdFiles))
	)
	for _, f := range c.UpdFiles {
		id, ok := files[f.Path]
		if !ok {
			excl := 0
			if f.ShouldExclude() {
				excl = 1
			}
			fID, err := zdb.InsertID(ctx, "file_id",
				`insert into files (repo_id, path, exclude) values (?, ?, ?)`,
				c.RepoID, f.Path, excl)
			if err != nil {
				return nil, nil, nil, err
			}
			id = fID
			files[f.Path] = id
		}

		// Binary files
		if f.Added == "-" {
			f.Added = "0"
		}
		if f.Removed == "-" {
			f.Removed = "0"
		}

		addedInt, err := strconv.ParseInt(f.Added, 10, 64)
		if err != nil {
			return nil, nil, nil, err
		}
		removedInt, err := strconv.ParseInt(f.Removed, 10, 64)
		if err != nil {
			return nil, nil, nil, err
		}

		commitFiles = append(commitFiles, id)
		added, removed = append(added, addedInt), append(removed, removedInt)
	}

	return commitFiles, added, removed, nil
}
