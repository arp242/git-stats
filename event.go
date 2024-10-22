package gitstats

import (
	"context"
	"time"

	"zgo.at/errors"
	"zgo.at/git-stats/db2"
	"zgo.at/zdb"
)

type Event struct {
	EventID int32     `db:"event_id,id" json:"-"`
	RepoID  int32     `db:"repo_id" json:"-"`
	Name    string    `db:"name" json:"name"`
	Date    time.Time `db:"date" json:"date"`
	Kind    EventKind `db:"kind" json:"kind"`
}

type EventKind byte

func (e EventKind) String() string {
	switch e {
	case 't':
		return "tag"
	case 'f':
		return "fork"
	case 'l':
		return "license change"
	case 'o':
		return "owner change"
	default:
		return "other"
	}
}

func (e *Event) Insert(ctx context.Context) error {
	err := db2.Insert(ctx, e)
	return errors.Wrap(err, "Event.Insert")
}

func (e *Event) Find(ctx context.Context) error {
	err := zdb.Get(ctx, e, `select * from events where repo_id=? and kind=? and name=?`, e.RepoID, e.Kind, e.Name)
	return errors.Wrap(err, "Event.Find")
}

type Events []Event

func (t *Events) List(ctx context.Context, repoID int32) error {
	err := zdb.Select(ctx, t, `select * from events where repo_id=? order by date asc`, repoID)
	return errors.Wrap(err, "Events.List")
}
