package db2_test

import (
	"context"
	"reflect"
	"testing"

	"zgo.at/git-stats/db2"
	"zgo.at/zdb"
	_ "zgo.at/zdb/drivers/pq"
	"zgo.at/zstd/ztest"
)

func TestInsert(t *testing.T) {
	zdb.RunTest(t, func(t *testing.T, ctx context.Context) {
		err := zdb.Exec(ctx, `create table tbl (id serial, str text, "NoTag" text)`)
		if err != nil {
			t.Fatal(err)
		}
		row := struct {
			ID        int    `db:"id,id"`
			Str       string `db:"str"`
			NoTag     string
			NoInsert  string `db:"other,noinsert"`
			Dash      string `db:"-"`
			unExport1 string
		}{0, "aaa", "bbb", "xxx", "xxx", "xxx"}
		db2.AddTable(row, "tbl")

		{ // Insert should work
			err := db2.Insert(ctx, &row)
			if err != nil {
				t.Fatal(err)
			}
			if row.ID != 1 {
				t.Fatalf("row.ID is not 1: %d", row.ID)
			}
			want := "id  str  NoTag\n1   aaa  bbb\n"
			if have := zdb.DumpString(ctx, "select * from tbl"); have != want {
				t.Fatal(have)
			}
		}

		{ // Fail if ID not zero value
			err := db2.Insert(ctx, &row)
			if !ztest.ErrorContains(err, "not zero value") {
				t.Fatal(err)
			}
			if row.ID != 1 {
				t.Fatalf("row.ID is not 1: %d", row.ID)
			}
			want := "id  str  NoTag\n1   aaa  bbb\n"
			if have := zdb.DumpString(ctx, "select * from tbl"); have != want {
				t.Fatal(have)
			}
		}

		{ // Fail on non-ptr
			err := db2.Insert(ctx, row)
			if !ztest.ErrorContains(err, "not a pointer") {
				t.Fatal(err)
			}
			want := "id  str  NoTag\n1   aaa  bbb\n"
			if have := zdb.DumpString(ctx, "select * from tbl"); have != want {
				t.Fatal(have)
			}
		}
	})
}

func TestUpdate(t *testing.T) {
	type Row struct {
		ID        int    `db:"id,id"`
		Str       string `db:"str"`
		NoTag     string
		NoInsert  string `db:"other,noinsert"`
		Dash      string `db:"-"`
		unExport1 string
	}
	db2.AddTable(Row{}, "tbl")

	zdb.RunTest(t, func(t *testing.T, ctx context.Context) {
		err := zdb.Exec(ctx, `create table tbl (id serial, str text, "NoTag" text)`)
		if err != nil {
			t.Fatal(err)
		}

		{ // ID is zero value
			old := Row{0, "aaa", "bbb", "xxx", "xxx", "xxx"}
			new := Row{}

			change, err := db2.Update(ctx, old, new)
			if !ztest.ErrorContains(err, "zero value") {
				t.Fatal(err)
			}
			if change != nil {
				t.Fatalf("change is not nil: %v", change)
			}
		}

		{ // Works
			old := Row{0, "aaa", "bbb", "xxx", "xxx", "xxx"}
			err := db2.Insert(ctx, &old)
			if err != nil {
				t.Fatal(err)
			}
			new := Row{0, "ddd", "bbb", "yyy", "yyy", "yyy"}

			change, err := db2.Update(ctx, old, new)
			if err != nil {
				t.Fatal(err)
			}
			wantC := map[string][2]any{
				"str": [2]any{"aaa", "ddd"},
			}
			if !reflect.DeepEqual(change, wantC) {
				t.Errorf("\nhave: %#v\nwant: %#v", change, wantC)
			}
			want := "id  str  NoTag\n1   ddd  bbb\n"
			if have := zdb.DumpString(ctx, "select * from tbl"); have != want {
				t.Fatal(have)
			}
		}
	})
}
