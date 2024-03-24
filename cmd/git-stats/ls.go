package main

import (
	"context"
	"fmt"

	"zgo.at/acidtab"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
)

func cmdLs(db zdb.DB, repoName string) error {
	var (
		repo gitstats.Repo
		ctx  = zdb.WithDB(context.Background(), db)
	)
	err := repo.ByName(ctx, repoName)
	if err != nil {
		return err
	}

	var files gitstats.FileStat
	err = files.List(zdb.WithDB(context.Background(), db), repo.ID)
	if err != nil {
		return err
	}

	t := acidtab.New("id", "commits", "path")
	t.Grow(len(files))
	for _, f := range files {
		t.Row(f.ID, f.NumCommits, f.Path)
	}
	fmt.Print(t.String())
	return nil
}
