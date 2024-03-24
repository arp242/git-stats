package main

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"zgo.at/acidtab"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	"zgo.at/zstd/zstring"
	"zgo.at/zstd/ztime"
)

func cmdAuthors(db zdb.DB, out io.Writer, repoName string) error {
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

	t := acidtab.New("commits", "added", "removed", "active", "author").Pad(" ").
		AlignCol(0, acidtab.Right).AlignCol(1, acidtab.Right).AlignCol(2, acidtab.Right)
	for _, r := range stat {
		t.Row(
			fmt.Sprintf("%s %2s%%", fmtNum(r.Commits), fmtFloat(r.CommitPerc)),
			fmt.Sprintf("%s %2s%%", fmtNum(r.Added), fmtFloat(r.AddedPerc)),
			fmt.Sprintf("%s %2s%%", fmtNum(r.Removed), fmtFloat(r.RemovedPerc)),
			r.First.Format("Jan 2006")+" – "+r.Last.Format("Jan 2006"),
			zstring.ElideLeft(r.Names.Join(", ")+", "+strings.Join(r.Domains, " "), 150))
	}
	out.Write([]byte(t.String()))
	return nil
}

func fmtFloat(n float32) string {
	if n < 0.95 {
		return fmt.Sprintf("%.1f", n)[1:]
	}
	return fmt.Sprintf("%0.0f", n)
}

func fmtNum(n int) string {
	sep := ','

	s := strconv.FormatInt(int64(n), 10)
	if len(s) < 4 {
		return s
	}

	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	var out []rune
	for i := range b {
		if i > 0 && i%3 == 0 && sep > 1 {
			out = append(out, sep)
		}
		out = append(out, rune(b[i]))
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}
