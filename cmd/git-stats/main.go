package main

import (
	"context"
	"fmt"
	"os"

	gitstats "zgo.at/git-stats"
	"zgo.at/zdb"
	_ "zgo.at/zdb/drivers/pq"
	"zgo.at/zli"
	"zgo.at/zstd/zfs"
	"zgo.at/zstd/zslice"
)

func main() {
	f := zli.NewFlags(os.Args)
	var (
		dbConnect = f.String("postgres+dbname=git-stats", "db")
		cache     = f.String("/tmp/git-stats", "cache")
		dev       = f.Bool(false, "dev")
		dbg       = f.Bool(false, "debug")
	)

	if zslice.ContainsAny(os.Args, "-h", "-help", "--help") {
		fmt.Print(usage)
		return
	}

	cmd, err := f.ShiftCommand("help", "update", "authors", "domains",
		"activity", "ls", "serve", "rm")
	zli.F(err)
	if cmd == "help" {
		fmt.Print(usage)
		return
	}
	zli.F(f.Parse(zli.AllowUnknown()))

	db := connectDB(dbConnect.String(), dev.Bool(), dbg.Bool())
	defer db.Close()

	switch cmd {
	case "ls":
		zli.F(f.Parse())
		zli.F(cmdLs(db, f.Shift()))
	case "rm":
		zli.F(f.Parse())
		zli.F(cmdRm(db, f.Shift()))
	case "update":
		var (
			noFetch = f.Bool(false, "no-fetch", "nofetch")
			keep    = f.Bool(false, "keep")
			name    = f.String("", "name")
		)
		zli.F(f.Parse())
		zli.F(cmdUpdate(db, cache.String(), f.Shift(), name.String(), noFetch.Bool(), keep.Bool()))
	case "authors":
		zli.F(f.Parse())
		zli.F(cmdAuthors(db, os.Stdout, f.Shift()))
	case "domains":
		zli.F(f.Parse())
		zli.F(cmdDomains(db, os.Stdout, f.Shift()))
	case "activity":
		zli.F(f.Parse())
		zli.F(cmdActivity(db, f.Shift()))
	case "serve":
		listen := f.String("127.0.0.1:8080", "listen")
		zli.F(f.Parse())
		zli.F(cmdServe(db, dev.Bool(), listen.String()))
	}
	zli.F(err)
}

func connectDB(connect string, dev, dbg bool) zdb.DB {
	fsys, err := zfs.EmbedOrDir(gitstats.DBFiles, "./db", dev)
	zli.F(err)
	db, err := zdb.Connect(context.Background(), zdb.ConnectOptions{
		Connect: connect,
		Files:   fsys,
		Create:  true,
	})
	zli.F(err)
	if dbg {
		db = zdb.NewLogDB(db, os.Stderr, zdb.DumpQuery, "")
	}
	return db
}
