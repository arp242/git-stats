package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zgo.at/errors"
	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	"zgo.at/zstd/zint"
	"zgo.at/zstd/zstring"
)

func getName(p string) string {
	if p[0] == '/' {
		return filepath.Base(filepath.Clean(p))
	}

	u, err := url.Parse(p)
	if err != nil {
		repo := strings.TrimPrefix(p, "git@")
		repo = strings.Replace(repo, ":", "/", 1)
		u, err = url.Parse(repo)
	}
	if err != nil {
		panic(err) // Should never happen really.
	}
	return path.Base(strings.TrimSuffix(path.Clean(u.Path), ".git"))
}

var mimeDec mime.WordDecoder

func cmdUpdate(db zdb.DB, cache, path, name string, noFetch, keep bool) error {
	if path == "" {
		return errors.New("need path")
	}

	err := os.MkdirAll(cache, 0o755)
	if err != nil {
		return err
	}

	var dir string
	err = zdb.TX(zdb.WithDB(context.Background(), db), func(ctx context.Context) error {
		var repo gitstats.Repo
		err := repo.Find(ctx, path)
		if zdb.ErrNoRows(err) {
			repo.Path = path
			repo.Name = name
			if name == "" {
				repo.Name = getName(path)
			}
			err = repo.Insert(ctx)
		}
		if err != nil {
			return err
		}

		known := findKnown(repo.Path)
		if known != nil {
			for _, e := range known {
				e.RepoID = repo.ID
				err := e.Find(ctx)
				if zdb.ErrNoRows(err) {
					err = e.Insert(ctx)
				}
				if err != nil {
					return err
				}
			}
		}

		dir = filepath.Join(cache, repo.Name)
		if noFetch {
			if _, err := os.Stat(dir); err != nil {
				return fmt.Errorf("-no-fetch was given, but %q doesn't exist", dir)
			}
		}
		if !noFetch {
			err := cloneOrUpdate(repo, dir, true)
			if err != nil {
				return errors.Wrap(err, "update")
			}
			fmt.Println()
		}

		return insertLog(ctx, dir, repo, true)
	})
	if err != nil {
		return errors.Wrap(err, "update")
	}

	if !keep {
		return os.RemoveAll(dir)
	}
	return nil
}

func emptydir(dir string) bool {
	if _, err := os.Stat(dir); err == nil {
		ls, err := os.ReadDir(dir)
		if err == nil && len(ls) == 0 {
			return false
		}
		return true
	}
	return false
}

func cloneOrUpdate(repo gitstats.Repo, dir string, verbose bool) error {
	exists := emptydir(dir)
	if !exists && !repo.Remote() {
		return fmt.Errorf("unable to clone %q: %q is not a remote URL", repo.Path, repo.Path)
	}

	args := make([]string, 0, 16)
	if exists {
		args = append(args, "-C", dir, "fetch", "--progress")
	} else {
		args = append(args, "clone", "--mirror", "--progress")
		if repo.LastCommit != nil {
			args = append(args, "--shallow-since", repo.LastCommitAt.Format(time.RFC3339))
		}
		args = append(args, repo.Path, dir)

		err := os.MkdirAll(dir, 0750)
		if err != nil {
			return errors.Wrap(err, "cloneOrUpdate")
		}
	}

	if verbose {
		fmt.Println("git-stats: run: git", strings.Join(args, " "))
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func fmtProg(n int) string {
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	if n < 1_000_000 {
		return s[:len(s)-3] + "," + s[len(s)-3:len(s)]
	}
	return s[:len(s)-6] + "," + s[len(s)-6:len(s)-3] + "," + s[len(s)-3:len(s)]
}

func insertLog(ctx context.Context, dir string, repo gitstats.Repo, verbose bool) error {
	/// Read existing data from DB.
	files, err := existingFiles(ctx, repo.ID)
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	names, emails, err := existingAuthors(ctx, repo.ID)
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}

	/// Read tags.
	var tt gitstats.Events
	err = tt.List(ctx, repo.ID)
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	existingTags := make(map[string]struct{}, len(tt))
	for _, t := range tt {
		existingTags[t.Name] = struct{}{}
	}

	tags, err := git("-C", dir, "tag", "--sort=authordate", "--format=%(authordate:short) %(refname:short)")
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	for _, l := range strings.Split(string(tags), "\n") {
		if l == "" {
			continue
		}
		d, t, ok := strings.Cut(l, " ")
		if !ok {
			return errors.Errorf("insertLog: parsing tag %q", l)
		}
		dd, err := time.Parse("2006-01-02", d)
		if !ok {
			return errors.Wrapf(err, "insertLog: parsing tag %q", l)
		}
		if _, ok := existingTags[t]; !ok {
			tag := gitstats.Event{
				RepoID: repo.ID,
				Name:   t,
				Date:   dd,
				Kind:   't',
			}
			err := tag.Insert(ctx)
			if err != nil {
				return errors.Wrap(err, "insertLog")
			}
		}
	}

	/// Read git log.
	out, t, err := readLog(repo, dir, verbose)
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	total := fmtProg(t)

	var (
		scan  = bufio.NewScanner(out)
		i     int
		start = time.Now()
		c     = gitstats.Commit{RepoID: repo.ID, UpdFiles: make([]gitstats.CommitFile, 0, 4)}
		n80   = time.Date(1980, 1, 1, 1, 0, 0, 0, time.UTC)
		n90   = time.Date(1990, 1, 1, 1, 0, 0, 0, time.UTC)
		bulk  = zdb.NewBulkInsert(ctx, "commits", []string{
			"repo_id", "date", "author_id", "files", "added", "removed", "hash", "subject",
			"added_sum", "removed_sum", "exclude"})
	)
	for scan.Scan() {
		l := scan.Text()

		switch {
		default: // Assume a --numstat file line
			var f gitstats.CommitFile
			f.Added, f.Removed, f.Path = zstring.Split3(l, "\t")
			f.Path = "/" + f.Path
			c.UpdFiles = append(c.UpdFiles, f)

		case l == "": // Ignore

		case strings.ContainsRune(l, 0): // Header
			if !c.Date.IsZero() { // Empty on first
				i++
				// Skip bots; rarely useful.
				//
				// Also skip commits before 1980; these are almost certainly
				// bogus. For example SDL (e14e0ef9) and angular.js (866346e1)
				// have commits from 1st of Jan 1970 completely out of place.
				// Probably a dev just having their computer clock wrong, or
				// something.
				//
				// Maybe there are some git repos with version control going
				// back to before mid-1980, but I've never seen them. SCCS was
				// first released in 1977. The oldest I can find is Emacs, with
				// the first commit being from Apr 1985.
				//
				// For Go, first four commits for Go are from 1972, 1974. These
				// are an intentional homage, or the like. Explicitly skip the
				// 1988 ones here.
				if strings.HasSuffix(c.Name, "[bot]") || c.Date.Before(n80) || (repo.Name == "go" && c.Date.Before(n90)) {
					c = gitstats.Commit{RepoID: repo.ID, UpdFiles: make([]gitstats.CommitFile, 0, 4)}
					continue
				}

				author, err := gitstats.FindOrInsertAuthor(ctx, c, names, emails)
				if err != nil {
					if bErr := bulk.Errors(); bErr != nil {
						return errors.Wrap(bErr, "insertLog")
					}
					return errors.Wrap(err, "insertLog")
				}
				commitFiles, added, removed, err := gitstats.FindOrInsertFile(ctx, c, files)
				if err != nil {
					if bErr := bulk.Errors(); bErr != nil {
						return errors.Wrap(bErr, "insertLog")
					}
					return errors.Wrap(err, "insertLog")
				}
				if i%10 == 0 {
					fmt.Printf("\rcommit %s / %s  ", fmtProg(i), total)
				}
				if len(commitFiles) > 0 {
					excl := 0
					if c.ShouldExclude() {
						excl = 1
					}
					if repo.FirstCommit == nil {
						c := c
						repo.FirstCommit, repo.FirstCommitAt = &c.Hash, &c.Date
					}
					bulk.Values(repo.ID, c.Date, author, "{"+zint.Join(commitFiles, ",")+"}",
						added, removed, c.Hash, c.Subject, added.Sum(), removed.Sum(), excl)
				}
				c = gitstats.Commit{RepoID: repo.ID, UpdFiles: make([]gitstats.CommitFile, 0, 4)}
			}

			// Ensure the name and subject are valid UTF-8 by converting it to a
			// []rune and back, which replaces invalid UTF-8 with the U+FFFD
			// replacment character. As far as I can tell, there isn't really
			// any good way to tell the encoding.
			//
			// For example https://github.com/scrapy/scrapy has commit 2b93b0a
			// with:
			// author Libor Nenad\xe1l <libor.nenadal@gmail.com> 1336222382 +0200
			// Which is ISO-8859-1, but there isn't any good way to know that.
			ll := strings.Split(l, "\x00")
			var h, d string
			d, c.Email, c.Name, h, c.Subject = ll[0], ll[1], string([]rune(ll[2])), ll[3],
				strings.TrimSpace(string([]rune(ll[4])))

			dd, err := strconv.ParseInt(d, 10, 64)
			if err != nil {
				return errors.Wrap(err, "insertLog")
			}
			c.Date, c.Hash = time.Unix(dd, 0), gitstats.NewHash(h)

			// Shouldn't be encoded, but sometimes they are, and it's fast
			// enough to just check here.
			if n, err := mimeDec.DecodeHeader(c.Name); err == nil {
				c.Name = n
			}
		}
	}
	fmt.Printf("\rcommit %s / %s  - finished; took: %ds\n", fmtNum(i), total, int(time.Now().Sub(start).Seconds()))
	err = bulk.Finish()
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	if scan.Err() != nil {
		return errors.Wrap(scan.Err(), "insertLog")
	}

	var last struct {
		Hash gitstats.Hash `db:"hash"`
		Date time.Time     `db:"date"`
	}
	err = zdb.Get(ctx, &last,
		`select hash, date from commits where repo_id=? order by date desc limit 1`,
		repo.ID)
	if err != nil {
		return errors.Wrap(err, "insertLog")
	}
	repo.LastCommit, repo.LastCommitAt = &last.Hash, &last.Date

	err = repo.Update(ctx)
	return errors.Wrap(err, "insertLog")
}

func git(args ...string) ([]byte, error) {
	fmt.Println("git", strings.Join(args, " "))
	out, err := exec.Command("git", args...).CombinedOutput()
	out = bytes.TrimRight(out, "\n")
	return out, err
}

func existingFiles(ctx context.Context, repoID int32) (map[string]int64, error) {
	var f gitstats.Files
	err := f.List(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return f.Map(), nil
}

func existingAuthors(ctx context.Context, repoID int32) (map[string]int64, map[string]int64, error) {
	var a gitstats.Authors
	err := a.List(ctx, repoID)
	if err != nil {
		return nil, nil, err
	}

	var (
		names  = make(map[string]int64, len(a))
		emails = make(map[string]int64, len(a))
	)
	for _, aa := range a {
		for _, n := range aa.Names {
			names[n] = aa.ID
		}
		for _, n := range aa.Emails {
			emails[n] = aa.ID
		}
	}
	return names, emails, nil
}

// --reverse (from old to new) is rather slow – it seems that git will get the
// full log before printing it. But it's nicer for the DB as the rows can be
// appended to in order. Otherwise we'd first insert:
//
//	2020-01-09
//	2020-01-08
//
// And then append on an update, resulting in:
//
//	2020-01-09
//	2020-01-08
//	2020-01-12
//	2020-01-11
func readLog(repo gitstats.Repo, dir string, verbose bool) (io.Reader, int, error) {
	outn, err := git("-C", dir, "rev-list", "--count", "--no-merges", "--no-renames", "HEAD")
	if err != nil {
		return nil, 0, err
	}
	n, err := strconv.Atoi(string(outn))
	if err != nil {
		return nil, 0, err
	}

	args := []string{"-C", dir, "log", "--reverse", "--no-merges", "--no-renames", "--numstat",
		"--format=format:%at%x00%aE%x00%aN%x00%H%x00%<(100,trunc)%s"}
	if repo.LastCommit != nil {
		args = append(args, repo.LastCommit.String()+"..")
	}

	cmd := exec.Command("git", args...)
	if verbose {
		fmt.Println(strings.Join(cmd.Args, " "))
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, 0, err
	}

	return out, n, nil
}
