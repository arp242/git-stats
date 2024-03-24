package main

import (
	"context"
	"fmt"
	"io"

	"zgo.at/acidtab"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	"zgo.at/zstd/ztime"
)

func cmdDomains(db zdb.DB, out io.Writer, repoName string) error {
	var (
		repo gitstats.Repo
		ctx  = zdb.WithDB(context.Background(), db)
	)
	err := repo.Find(ctx, repoName)
	if err != nil {
		return err
	}

	var stat gitstats.AuthorStats
	err = stat.List(ctx, repo.ID, "", ztime.Range{})
	if err != nil {
		return err
	}

	n, err := gitstats.CountCommits(ctx, repo.ID, ztime.Range{})
	if err != nil {
		return err
	}
	cnt := float32(n)

	t := acidtab.New("commits", "%", "domain").
		AlignCol(0, acidtab.Right).AlignCol(1, acidtab.Right)
	for _, r := range stat.Domains() {
		t.Row(
			fmtNum(r.Count),
			fmt.Sprintf("%.1f%%", float32(r.Count)/cnt*100),
			r.Domain)
	}
	out.Write([]byte(t.String()))
	return nil
}
