// TODO: most of this should probably be integrated in zdb; just not entirely
// sure yet about the API for that.

package db2

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"zgo.at/zdb"
	"zgo.at/zstd/zreflect"
)

var (
	tables   = make(map[reflect.Type]string)
	tablesMu sync.RWMutex
)

func AddTable(t any, tbl string) {
	typ := reflect.TypeOf(t)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	tablesMu.Lock()
	defer tablesMu.Unlock()
	if have, ok := tables[typ]; ok && have != tbl {
		panic(fmt.Sprintf("zdb.AddTable: type %T is already registered to table %q", t, have))
	}
	tables[typ] = tbl
}

func tblName(t any) string {
	typ := reflect.TypeOf(t)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	tablesMu.RLock()
	defer tablesMu.RUnlock()
	tbl, ok := tables[typ]
	if !ok {
		panic(fmt.Sprintf(`type %T is not registered; use AddTable()`, t))
	}
	return tbl
}

// Insert all struct fields of t in the table tbl.
//
// Column names are taken from the db tag. Fields with the db tag set to "-" or
// that have the ",noinsert" option will be skipped.
//
// If a field has the ",id" option it will be set.
func Insert(ctx context.Context, t any, extraSQL ...string) error {
	tbl := tblName(t)
	cols, params, opts := fields(t)

	// Get the ID column, if any
	var (
		idCol     any
		idColName string
	)
	for i, o := range opts {
		if slices.Contains(o, "id") {
			types := reflect.TypeOf(t)
			if types.Kind() != reflect.Ptr {
				return errors.New("zdb.Insert: type is not a pointer")
			}
			types = types.Elem()

			id := reflect.ValueOf(t).Elem().Field(i)
			if !id.IsZero() {
				return fmt.Errorf(`zdb.Insert: id field %q is not zero value but "%v"`, "f.Name", id.Interface())
			}

			idCol = id.Addr().Interface()
			idColName = cols[i]

			params = append(params[:i], params[i+1:]...)
			cols = append(cols[:i], cols[i+1:]...)
			break
		}
	}

	for i := range cols {
		cols[i] = quoteIdentifier(cols[i])
	}
	q := fmt.Sprintf(`insert into %s (%s) values (?) %s`,
		quoteIdentifier(tbl),
		strings.Join(cols, ", "),
		strings.Join(extraSQL, " "))
	var err error
	if idColName == "" {
		err = zdb.Exec(ctx, q, params)
	} else {
		q += " returning " + quoteIdentifier(idColName)
		err = zdb.Get(ctx, idCol, q, params)
	}
	if err != nil {
		return fmt.Errorf("zdb.Insert: %w\n%s", err, q)
	}
	return nil
}

// Update all fields that are changed between old and new.
//
// No SQL is run if old and new are identical.
//
// Returns a map of changes, as [2]any{oldValue, newValue}.
func Update(ctx context.Context, old, new any) (map[string][2]any, error) {
	if old != nil && reflect.TypeOf(old) != reflect.TypeOf(new) {
		return nil, errors.New("zdb.Update: old and new are not the same types")
	}

	tbl := tblName(new)
	cols, newVals, opts := fields(new)
	var oldVals []any
	if old != nil {
		oldVals = values(old)
	}

	var (
		set        []string
		updParams  []any
		changes    map[string][2]any
		where      string
		whereParam any
	)
	for i := range cols {
		if slices.Contains(opts[i], "id") {
			var vv any
			if old == nil {
				vv = newVals[i]
			} else {
				vv = oldVals[i]
			}

			if reflect.ValueOf(vv).IsZero() {
				return nil, errors.New("zdb.Update: ID column is zero value")
			}
			where = fmt.Sprintf(`%s = ?`, quoteIdentifier(cols[i]))
			whereParam = vv
			continue
		}

		var eq bool
		if old != nil {
			switch reflect.TypeOf(oldVals[i]).Kind() {
			case reflect.Ptr:
				if v := reflect.ValueOf(oldVals[i]); !v.IsNil() {
					oldVals[i] = v.Elem().Interface()
				}
				if v := reflect.ValueOf(newVals[i]); !v.IsNil() {
					newVals[i] = v.Elem().Interface()
				}
				fallthrough
			case reflect.Slice, reflect.Map:
				eq = reflect.DeepEqual(oldVals[i], newVals[i])
			default:
				eq = oldVals[i] == newVals[i]
			}
		}

		if !eq {
			set = append(set, quoteIdentifier(cols[i])+` = ?`)
			updParams = append(updParams, newVals[i])
			if changes == nil {
				changes = make(map[string][2]any)
			}
			if old == nil {
				changes[cols[i]] = [2]any{nil, newVals[i]}
			} else {
				changes[cols[i]] = [2]any{oldVals[i], newVals[i]}
			}
		}
	}
	if len(set) > 0 {
		q := fmt.Sprintf(`update %s set %s where %s`,
			quoteIdentifier(tbl), strings.Join(set, ", "), where)
		err := zdb.Exec(ctx, q, append(updParams, whereParam)...)
		if err != nil {
			return changes, fmt.Errorf("zdb.Update: %w", err)
		}
	}
	return changes, nil
}

func fields(t any) ([]string, []any, [][]string) { return zreflect.Fields(t, "db", "noinsert") }
func names(t any) []string                       { return zreflect.Names(t, "db", "noinsert") }
func values(t any) []any                         { return zreflect.Values(t, "db", "noinsert") }
func namesQuoted(t any) []string {
	names := names(t)
	for i := range names {
		names[i] = quoteIdentifier(names[i])
	}
	return names
}
func quoteIdentifier(ident string) string {
	var b strings.Builder
	b.Grow(len(ident) + 2)
	b.WriteByte('"')
	for _, c := range ident {
		if c == '"' {
			b.WriteByte('"')
		}
		b.WriteRune(c)
	}
	b.WriteByte('"')
	return b.String()
}
