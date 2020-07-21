package pgUtils

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

type Direction string

const (
	Asc  Direction = "ASC"
	Desc Direction = "DESC"
)

type Sort interface {
	String(string) string
}

type SimpleSort string

func (s SimpleSort) String(key string) string {
	return fmt.Sprintf("%s %s", s, getDirection(key))
}

type ExtendedSort struct {
	Field          SimpleSort
	SecondarySorts []string
}

func (s ExtendedSort) String(key string) string {
	primary := s.Field.String(key)
	if len(s.SecondarySorts) == 0 {
		return primary
	}
	return primary + ", " + strings.Join(s.SecondarySorts, ", ")
}

func getDirection(key string) Direction {
	if key[0] == '-' {
		return Desc
	}
	return Asc
}

func getSortKey(key string) string {
	if key[0] == '-' {
		return strings.ToLower(key[1:])
	}
	return strings.ToLower(key)
}

func ApplySortStrings(xs []string, query *squirrel.SelectBuilder, sorts map[string]string) {
	p := make(map[string]Sort)
	for k, v := range sorts {
		p[k] = SimpleSort(v)
	}
	ApplySort(xs, query, p)
}

func ApplySort(activeSorts []string, query *squirrel.SelectBuilder, possibleSorts map[string]Sort) {
	for _, s := range activeSorts {
		if len(s) == 0 {
			continue
		}
		key := getSortKey(s)
		v, ok := possibleSorts[key]
		if !ok {
			continue
		}
		*query = query.OrderBy(v.String(s))
	}
}