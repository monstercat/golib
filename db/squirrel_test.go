package dbUtil

import (
	"strings"
	"testing"
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
