package gitstats

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"html/template"
	"path"
	"strconv"
	"strings"

	"zgo.at/zstd/zint"
)

var quote = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func isSpace(ch byte) bool {
	// see array_isspace:
	// https://github.com/postgres/postgres/blob/master/src/backend/utils/adt/arrayfuncs.c
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

type Strings []string

func (l Strings) String() string         { return strings.Join(l, ", ") }
func (l Strings) Join(sep string) string { return strings.Join(l, sep) }
func (l Strings) Value() (driver.Value, error) {
	for i := range l {
		if l[i] == "" || l[i] == "null" ||
			isSpace(l[i][0]) || isSpace(l[i][len(l[i])-1]) ||
			strings.ContainsAny(l[i], `{},"\`) {
			l[i] = `"` + quote.Replace(l[i]) + `"`
		}
	}
	return `{` + strings.Join(l, `,`) + `}`, nil
}
func (l *Strings) Scan(v any) error {
	if v == nil {
		return nil
	}

	var val []byte
	switch vv := v.(type) {
	case []byte:
		val = vv
	case string:
		val = []byte(vv)
	default:
		return fmt.Errorf("unsupported type: %T", vv)
	}

	if len(val) == 2 && val[0] == '{' && val[1] == '}' {
		*l = []string{}
		return nil
	}

	// TODO: wrong! Need to properly parse, deal with quotes, etc.
	// This is why Ulrich Drepper errors out, because e.g. commit
	// 5ab7f7c12a6c0 has file 1619242 with a comma:
	// /localedata/fr_CA,2.13.in
	//
	var (
		q, bs bool
		cur   = make([]rune, 0, 16)
		strs  = make([]string, 0, 8)
	)
	for _, c := range []rune(string(val[1 : len(val)-1])) {
		switch {
		case !q && c == ',':
			if len(cur) > 0 {
				strs = append(strs, string(cur))
			}
			cur = cur[:0]
		case !bs && c == '"':
			if q {
				strs = append(strs, string(cur))
				cur = cur[:0]
			}
			q = !q
		case c == '\\':
			bs = true
		default:
			cur = append(cur, c)
			bs = false
		}
	}
	if len(cur) > 0 {
		strs = append(strs, string(cur))
	}

	//split := bytes.Split(val[1:len(val)-1], []byte{','})
	//strs := make([]string, 0, len(split))
	//for _, s := range split {
	//	//s = strings.TrimSpace(s)
	//	//if len(s) == 0 {
	//	//	continue
	//	//}
	//	strs = append(strs, strings.Trim(string(s), `"`))
	//}
	*l = strs
	return nil
}

type Ints []int64

func (l Ints) Sum() int64 {
	var s int64
	for _, n := range l {
		s += n
	}
	return s
}

func (l Ints) String() string { return zint.Join(l, ", ") }
func (l Ints) Value() (driver.Value, error) {
	return "{" + zint.Join(l, ",") + "}", nil
}
func (l *Ints) Scan(v any) error {
	if v == nil {
		return nil
	}

	var val []byte
	switch vv := v.(type) {
	case []byte:
		val = vv
	case string:
		val = []byte(vv)
	default:
		return fmt.Errorf("unsupported type: %T", vv)
	}
	if len(val) == 2 && val[0] == '{' && val[1] == '}' {
		*l = []int64{}
		return nil
	}

	split := bytes.Split(val[1:len(val)-1], []byte{','})
	strs := make([]int64, 0, len(split))
	for _, s := range split {
		i, err := strconv.ParseInt(string(s), 10, 64)
		if err != nil {
			return err
		}
		strs = append(strs, i)
	}
	*l = strs
	return nil
}

type Hash [20]byte

func NewHash(s string) Hash {
	if len(s) != 40 {
		panic(fmt.Sprintf("wrong sha: %q", s))
	}
	var h Hash
	for i := 0; i < 40; i += 2 {
		n, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			panic(fmt.Sprintf("wrong sha: %q: %s", h, err))
		}
		h[i/2] = byte(n)
	}
	return h
}

func (h Hash) Short() string {
	// TODO: git is a bit smarter about this, using a short hash length
	// proportional to the number of commits or some such. Look up how this
	// works and duplicate.
	b := new(strings.Builder)
	for _, hh := range h[:4] {
		fmt.Fprintf(b, "%02x", hh)
	}
	return b.String()
}

func (h Hash) Link(repoURL string) template.HTML {
	l := h.URL(repoURL)
	if l == "" {
		return template.HTML(h.Short())
	}
	return template.HTML(`<a href="` + string(l) + `">` + h.Short() + `</a>`)
}

func (h Hash) URL(repoURL string) template.URL {
	switch {
	case strings.Contains(repoURL, "gitlab"):
		return template.URL(repoURL + "/-/commit/" + h.String())
	case strings.Contains(repoURL, "github.org") || strings.Contains(repoURL, "codeberg.org"):
		return template.URL(repoURL + "/commit/" + h.String())
	case strings.Contains(repoURL, "savannah.gnu.org"):
		n := path.Base(repoURL)
		return template.URL(fmt.Sprintf("https://git.savannah.gnu.org/cgit/%s.git/commit/?id=%s", n, h))
	default:
		return ""
	}
}

func (h Hash) String() string {
	b := new(strings.Builder)
	for _, hh := range h {
		fmt.Fprintf(b, "%02x", hh)
	}
	return b.String()
}

func (h *Hash) Scan(v any) error {
	switch vv := v.(type) {
	case string:
		*h = NewHash(vv)
	case []byte:
		*h = [20]byte(vv)
	default:
		return fmt.Errorf("Hash.Scan: unsupported type %T: %[1]q", vv)
	}
	return nil
}

func (h Hash) Value() (driver.Value, error) {
	return h[:], nil
}
