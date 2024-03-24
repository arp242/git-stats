package main

import (
	"context"

	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
)

func cmdRm(db zdb.DB, repoName string) error {
	var (
		repo gitstats.Repo
		ctx  = zdb.WithDB(context.Background(), db)
	)
	err := repo.ByName(ctx, repoName)
	if err != nil {
		return err
	}

	tbl := []string{"authors", "files", "commits", "repos"}
	return zdb.TX(ctx, func(ctx context.Context) error {
		for _, t := range tbl {
			if err := zdb.Exec(ctx, `delete from `+t+` where repo_id=?`, repo.ID); err != nil {
				return err
			}
		}
		return nil
	})
}
