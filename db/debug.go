package dbutil

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

func DebugQuery(query squirrel.Sqlizer) {
	DebugQueryPieces(query.ToSql())
}

func DebugQueryPieces(sql string, args []interface{}, err error) {
	fmt.Println("sql", sql)
	for i, arg := range args {
		fmt.Println("args", i+1, arg)
	}
	fmt.Println("err", err)
}
