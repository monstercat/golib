package postgres

import (
	"fmt"
	"strings"
)

// ColumnTranslatorGenerator returns a generator that causes a column name to be
// returned with a provided table. The returned value will be [tbl].[col]. If
// the provided FT field does not match to a column, simply return "".
//
// For table columns that do *not* match to the base table, e.g., because they
// are the result of a joined table, simply return "" and wrap with a different
// column translator function.
//
// For example,
//
// ```
//
//	MergeColumnTranslators[FT](
//	   ColumnTranslatorGenerator[FT](tbl, columnTranslatorFunc),
//	)
//
// ```
func ColumnTranslatorGenerator[FT ~string](
	tbl string,
	colName func(t FT) string,
) ColumnTranslator[FT] {
	return func(f FT) string {
		col := colName(f)
		if col == "" {
			return ""
		}
		return T(tbl, col)
	}
}

// MergeColumnTranslators merges column translators. It sequentially calls each
// translator until one which does not return ""
func MergeColumnTranslators[FT ~string](tls ...ColumnTranslator[FT]) ColumnTranslator[FT] {
	return func(col FT) string {
		for _, tl := range tls {
			if c := tl(col); c != "" {
				return c
			}
		}
		return ""
	}
}

// T quickly converts two strings representing a table and a field to a single
// one.
func T(table, field string) string {
	if len(table) == 0 {
		return field
	}
	if table[len(table)-1] == '.' {
		table = table[:len(table)-1]
	}
	if len(table) == 0 {
		return field
	}
	return fmt.Sprintf("%s.%s", table, field)
}

// UPPER takes a string and returns the input string wrapped in UPPER function.
// The returned string is in the format "UPPER(inputString)".
// It is used to convert the input string to uppercase in SQL queries.
func UPPER(s string) string {
	return fmt.Sprintf("UPPER(%s)", s)
}

// Coalesce returns the provided SQL wrapped with a coalesce.
func Coalesce(sql ...string) string {
	return fmt.Sprintf("COALESCE(%s)", strings.Join(sql, ","))
}

// NullIf returns the provided SQL statements to convert a result to null
func NullIf(sql1, sql2 string) string {
	return fmt.Sprintf("NULLIF(%s, %s)", sql1, sql2)
}

// CastString casts the column to a string.
func CastString(sql string) string {
	return fmt.Sprintf("(%s)::TEXT", sql)
}

// CoalesceString coalesces as a string.
func CoalesceString(sql ...string) string {
	if len(sql) == 1 {
		return Coalesce(sql[0]+"::TEXT", "''")
	}
	return Coalesce(Coalesce(sql...)+"::TEXT", "''")
}

// CoalesceNumber coalesces into a number
func CoalesceNumber(sql ...string) string {
	return Coalesce(append(sql, "0")...)
}

// CoalesceFalse coalesces to false.
func CoalesceFalse(sql ...string) string {
	return Coalesce(append(sql, "false")...)
}

// As creates an alias for a column.
func As(sql string, as string) string {
	return fmt.Sprintf("(%s) as %s", sql, as)
}

// Or returns a string representing an OR statement.
func Or(sql ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(sql, " OR "))
}

// Sum returns a column for a SUM() function.
func Sum(sql string) string {
	return fmt.Sprintf("SUM(%s)", sql)
}

// Count returns a column for a COUNT() function.
func Count(sql string) string {
	return fmt.Sprintf("COUNT(%s)", sql)
}

// Min returns a column for a MIN() function.
func Min(sql string) string {
	return fmt.Sprintf("MIN(%s)", sql)
}

// Max returns a column for a MAX() function.
func Max(sql string) string {
	return fmt.Sprintf("MAX(%s)", sql)
}

// Greatest returns a column for a Greatest() function.
func Greatest(sql ...string) string {
	expr := strings.Join(sql, ",")
	return fmt.Sprintf("GREATEST(%s)", expr)
}

// CountDistinct returns a column for a COUNT(DISTINCT) function.
func CountDistinct(sql string) string {
	return fmt.Sprintf("COUNT(DISTINCT %s)", sql)
}

// WhenThen returns a WHEN THEN statement. This should be used as a parameter for
// Case
func WhenThen(when string, then string) string {
	return fmt.Sprintf("WHEN %s THEN %s", when, then)
}

// Else returns an ELSE statement. This should be used as a parameter for Case
func Else(e string) string {
	return fmt.Sprintf("ELSE %s", e)
}

// Case returns a CASE statement. Pass in the results from WhenThen and Else
// in order. Else should be the last case.
func Case(when ...string) string {
	var str strings.Builder
	str.WriteString("CASE ")
	str.WriteString(strings.Join(when, " "))
	str.WriteString(" END")
	return str.String()
}
