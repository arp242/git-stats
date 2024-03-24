package main

import (
	"context"
	"fmt"

	"zgo.at/acidtab"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	"zgo.at/zstd/ztime"
)

func cmdActivity(db zdb.DB, repoName string) error {
	var (
		repo gitstats.Repo
		ctx  = zdb.WithDB(context.Background(), db)
	)
	err := repo.Find(ctx, repoName)
	if err != nil {
		return err
	}

	var act gitstats.CommitStats
	err = act.List(zdb.WithDB(context.Background(), db), repo.ID, ztime.Range{}, 0, true)
	if err != nil {
		return err
	}
	t := acidtab.New("month", "commits", "addded", "removed").
		AlignCol(1, acidtab.Right).
		AlignCol(2, acidtab.Right).
		AlignCol(3, acidtab.Right).
		FormatColFunc(1, acidtab.FormatAsNum()).
		FormatColFunc(2, acidtab.FormatAsNum()).
		FormatColFunc(3, acidtab.FormatAsNum())
	t.Grow(len(act))
	for _, a := range act {
		t.Row(a.Date.Format("2006-01"), a.Commits, a.Added, a.Removed)
	}
	fmt.Print(t.String())
	return nil
}
