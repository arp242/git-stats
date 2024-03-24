package gitstats

import (
	"embed"

	"zgo.at/git-stats/db2"
)

//go:embed db/*
var DBFiles embed.FS

//go:embed tpl/*
var TplFiles embed.FS

func init() {
	db2.AddTable(Repo{}, "repos")
	db2.AddTable(Author{}, "authors")
	db2.AddTable(File{}, "files")
	db2.AddTable(Commit{}, "commits")
	db2.AddTable(Event{}, "events")
}
