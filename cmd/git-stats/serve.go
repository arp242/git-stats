package main

import (
	"cmp"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/teamwork/reload"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	"zgo.at/zhttp"
	"zgo.at/zhttp/mware"
	"zgo.at/zli"
	"zgo.at/zstd/zfs"
	"zgo.at/zstd/ztime"
	"zgo.at/ztpl"
	"zgo.at/ztpl/tplfunc"
)

func cmdServe(db zdb.DB, dev bool, listen string) error {
	tplfunc.Add("hash", func(h string) gitstats.Hash {
		return gitstats.NewHash(h[3:])
	})
	tplfunc.Add("reverse", func(s []gitstats.Event) []gitstats.Event {
		ss := slices.Clone(s)
		slices.Reverse(ss)
		return ss
	})

	fsys, err := zfs.EmbedOrDir(gitstats.TplFiles, "./tpl", dev)
	if err != nil {
		return err
	}
	err = ztpl.Init(fsys)
	if err != nil {
		return err
	}

	if dev {
		go func() {
			err := reload.Do(
				func(s string, args ...any) { fmt.Printf(s+"\n", args...) },
				reload.Dir("./tpl", func() { ztpl.Reload("./tpl") }),
			)
			zli.F(err)
		}()
	}

	rdy, err := zhttp.Serve(0, nil, &http.Server{
		Addr:    listen,
		Handler: newHandler(db, dev),
	})
	if err != nil {
		return err
	}
	<-rdy
	fmt.Println("serving on", listen)
	<-rdy
	return nil
}

func addctx(db zdb.DB) zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			*r = *r.WithContext(zdb.WithDB(r.Context(), db))
			return next(w, r)
		}
	}
}

func newHandler(db zdb.DB, dev bool) http.Handler {
	mw := []zhttp.Middleware{
		mware.Unpanic(),
		addctx(db),
		mware.RequestLog(nil),
	}
	if dev {
		mw = append(mw, mware.NoCache())
	} else {
		// Cache up to a month in Varnish; none of this changes all that fast.
		mw = append(mw, mware.Headers(http.Header{"Cache-Control": []string{"public,max-age=2592000"}}))
	}
	m := zhttp.NewServeMux()
	m.HandleFunc("GET /", index, mw...)
	m.HandleFunc("GET /{repo}/", repo, mw...)
	m.HandleFunc("GET /{repo}/{author_id}/", author, mw...)
	return m
}

func scripts() []template.JS {
	_, err := os.Stat("./tpl/zcript.js")
	dev := err == nil // TODO: should use -dev
	fsys, err := zfs.EmbedOrDir(gitstats.TplFiles, "./tpl", dev)
	if err != nil {
		panic(err)
	}

	ls, err := fs.ReadDir(fsys, ".")
	if err != nil {
		panic(err)
	}
	js := make([]template.JS, 0, 4)
	for _, f := range ls {
		if strings.HasSuffix(f.Name(), ".js") {
			d, err := fs.ReadFile(fsys, f.Name())
			if err != nil {
				panic(err)
			}
			js = append(js, template.JS(d))
		}
	}

	return js
}

func index(w http.ResponseWriter, r *http.Request) error {
	var repos gitstats.Repos
	err := repos.List(r.Context())
	if err != nil {
		return err
	}

	return zhttp.Template(w, "index.gohtml", struct {
		Scripts []template.JS
		Repos   gitstats.Repos
	}{scripts(), repos})
}

func repo(w http.ResponseWriter, r *http.Request) error {
	var args struct {
		Start    string `json:"start"`
		End      string `json:"end"`
		Selected string `json:"sel"`
	}
	if _, err := zhttp.Decode(r, &args); err != nil {
		return err
	}

	var repo gitstats.Repo
	err := repo.ByName(r.Context(), r.PathValue("repo"))
	if err != nil {
		return err
	}

	var rng ztime.Range
	if t, err := time.Parse("2006-01-02", args.Start); err == nil {
		rng.Start = t
	}
	if t, err := time.Parse("2006-01-02", args.End); err == nil {
		rng.End = t
	}

	var (
		wg              sync.WaitGroup
		stat            gitstats.AuthorStats
		act             gitstats.CommitStats
		statErr, actErr error
	)
	wg.Add(2)
	go func() { defer wg.Done(); statErr = stat.List(r.Context(), repo.ID, "", rng) }()
	go func() { defer wg.Done(); actErr = act.List(r.Context(), repo.ID, rng, 0, false) }()
	wg.Wait()
	if statErr != nil {
		return statErr
	}
	if actErr != nil {
		return actErr
	}

	now := ztime.Time{time.Now()}
	month, halfYear, year, fiveYear, decade := now.AddPeriod(-1, ztime.Month), now.AddPeriod(-1, ztime.HalfYear),
		now.AddPeriod(-1, ztime.Year), now.AddPeriod(-5, ztime.Year), now.AddPeriod(-10, ztime.Year)

	n, err := gitstats.CountCommits(r.Context(), repo.ID, rng)
	if err != nil {
		return err
	}

	ev := gitstats.Events{}
	err = ev.List(r.Context(), repo.ID)
	if err != nil {
		return err
	}

	if rng.Start.IsZero() && repo.FirstCommitAt != nil {
		rng.Start = *repo.FirstCommitAt
	}
	if rng.End.IsZero() && repo.LastCommitAt != nil {
		rng.End = time.Now()
	}

	fillblank(&act, rng)

	return zhttp.Template(w, "repo.gohtml", struct {
		Scripts                                  []template.JS
		Repo                                     gitstats.Repo
		AuthorStats                              gitstats.AuthorStats
		Activity                                 gitstats.CommitStats
		Events                                   gitstats.Events
		NumCommits                               int
		Range                                    ztime.Range
		Selected                                 string
		Month, HalfYear, Year, FiveYears, Decade time.Time
	}{scripts(), repo, stat, act, ev, n, rng, args.Selected,
		month.Time, halfYear.Time, year.Time, fiveYear.Time, decade.Time})
}

func author(w http.ResponseWriter, r *http.Request) error {
	var args struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if _, err := zhttp.Decode(r, &args); err != nil {
		return err
	}

	var repo gitstats.Repo
	err := repo.ByName(r.Context(), r.PathValue("repo"))
	if err != nil {
		return err
	}

	authorID, err := strconv.ParseInt(r.PathValue("author_id"), 10, 64)
	if err != nil {
		return err
	}

	var author gitstats.Author
	err = author.ByID(r.Context(), repo.ID, authorID)
	if err != nil {
		return err
	}

	var rng ztime.Range
	if t, err := time.Parse("2006-01-02", args.Start); err == nil {
		rng.Start = t
	}
	if t, err := time.Parse("2006-01-02", args.End); err == nil {
		rng.End = t
	}

	var (
		stat                    gitstats.AuthorStat
		commits                 gitstats.Commits
		act                     gitstats.CommitStats
		wg                      sync.WaitGroup
		statErr, actErr, comErr error
	)
	wg.Add(3)
	go func() { defer wg.Done(); statErr = stat.ByID(r.Context(), repo.ID, authorID, rng) }()
	go func() { defer wg.Done(); actErr = act.List(r.Context(), repo.ID, rng, authorID, false) }()
	go func() { defer wg.Done(); comErr = commits.ByAuthor(r.Context(), repo.ID, authorID, rng) }()
	wg.Wait()
	if statErr != nil {
		return statErr
	}
	if actErr != nil {
		return statErr
	}
	if comErr != nil {
		return comErr
	}

	now := ztime.Time{time.Now()}
	month, halfYear, year, fiveYear, decade := now.AddPeriod(-1, ztime.Month), now.AddPeriod(-1, ztime.HalfYear),
		now.AddPeriod(-1, ztime.Year), now.AddPeriod(-5, ztime.Year), now.AddPeriod(-10, ztime.Year)

	if rng.Start.IsZero() && repo.FirstCommitAt != nil {
		rng.Start = *repo.FirstCommitAt
	}
	if rng.End.IsZero() && repo.LastCommitAt != nil {
		rng.End = time.Now()
	}

	fillblank(&act, rng)

	type file struct {
		Path                    string
		Commits, Added, Removed int64
	}

	all := make([]file, 0, 4)
	for _, c := range commits {
	file:
		for i, f := range c.Files {
			for j, ff := range all {
				if f == ff.Path {
					all[j].Commits++
					all[j].Added += c.Added[i]
					all[j].Removed += c.Removed[i]
					continue file
				}
			}

			all = append(all, file{
				Path:    f,
				Commits: 1,
				Added:   c.Added[i],
				Removed: c.Removed[i],
			})
		}
	}
	slices.SortFunc(all, func(a, b file) int {
		return cmp.Compare(b.Commits, a.Commits)
	})

	return zhttp.Template(w, "author.gohtml", struct {
		Scripts                                  []template.JS
		Repo                                     gitstats.Repo
		Author                                   gitstats.Author
		AuthorStat                               gitstats.AuthorStat
		Activity                                 gitstats.CommitStats
		Commits                                  gitstats.Commits
		Range                                    ztime.Range
		AllFiles                                 []file
		Month, HalfYear, Year, FiveYears, Decade time.Time
	}{scripts(), repo, author, stat, act, commits, rng, all,
		month.Time, halfYear.Time, year.Time, fiveYear.Time, decade.Time})
}

// Fill in blank days.
func fillblank(actptr *gitstats.CommitStats, rng ztime.Range) {
	act := *actptr
	if len(act) == 0 {
		return
	}
	var (
		newact = make(gitstats.CommitStats, 0, len(act))
		day    = act[0].Date.Add(-24 * time.Hour)
		endFmt = act[len(act)-1].Date.Format("2006-01-02")
		i      int
	)
	if !rng.Start.IsZero() {
		day = rng.Start.Add(-24 * time.Hour)
	}
	if !rng.End.IsZero() {
		endFmt = rng.End.Format("2006-01-02")
	}
	for {
		day = day.Add(24 * time.Hour)
		dayFmt := day.Format("2006-01-02")

		if len(act)-1 >= i && dayFmt == act[i].Date.Format("2006-01-02") {
			newact = append(newact, act[i])
			i++
		} else {
			newact = append(newact, gitstats.CommitStat{Date: day})
		}

		if dayFmt == endFmt {
			break
		}
	}
	*actptr = newact
}
