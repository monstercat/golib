package dbUtil

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
)

func TestILike_ToSql(t *testing.T) {

	x := ILikeOr{
		"id":     "abs",
		"isrc":   "ca6",
		"cannot": nil,
		"array":  []int{1, 2, 3},
	}

	sql, args, err := x.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if len(args) != 2 {
		t.Fatal("Result should have two args.")
	}

	if strings.Index(sql, "id ILIKE") == -1 {
		t.Fatal("id should be part of the resultant sql")
	}
	if strings.Index(sql, "isrc ILIKE") == -1 {
		t.Fatal("isrc should be part of the resultant sql")
	}

	if args[0] == "abs" && args[1] == "ca6" || args[0] == "ca6" && args[1] == "abs" {
		return
	}
	t.Fatalf("incorrect arguments provided. Should be 'abs' and 'ca6. Got %v",args)
}

func TestUnion(t *testing.T) {
	qry, err := Union(
		squirrel.Select("a", "b", "c").From("table 1"),
		squirrel.Select("a", "b", "c").From("table 2"),
		squirrel.Select("a", "b", "c").From("table 3"),
	)
	if err != nil {
		t.Fatal(err)
	}

	sql, _, err := qry.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if sql != "(SELECT a, b, c FROM table 2) UNION (SELECT a, b, c FROM table 3) UNION SELECT a, b, c FROM table 1" {
		t.Fatal("unexpected sql")
	}
}