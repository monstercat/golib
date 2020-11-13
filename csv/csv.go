package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/trubitsyn/go-zero-width"
)

type Iterator func(row []string, line int) error
type MapIterator func(row map[string]string, line int) error
type MapIteratorGroupErrors func(row map[string]string, line int, errs *MultipleError)

func IterateCsv(r *csv.Reader, lambda Iterator) error {
	i := 1
	for {
		// Read next row.
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := lambda(row, i); err != nil {
			return err
		}

		i++
	}
	return nil
}

// Iterates through the CSV
func IterateCsvMap(r *csv.Reader, expectedHeaders []string, lambda MapIterator) error {
	// Read the first row and parse it.
	headerRow, err := r.Read()
	if err == io.EOF {
		// Done! no data!
		return nil
	}
	minRowCols, err := CheckCsvHeaders(headerRow, expectedHeaders)
	if err != nil {
		return err
	}

	// From now on, its the second row. So, we need to add 1 to line
	return IterateCsv(r, func(row []string, line int) error {
		i := line + 1

		// Ensure correct # of columns.
		if len(row) < minRowCols {
			return ColumnMismatchError{
				Line: i,
				Expected: minRowCols,
				Got: len(row),
			}
		}
		// Create a map from the column headers.
		// We can pass this back to the iterative function.
		data := map[string]string{}
		for i, c := range row {
			data[CleanCell(headerRow[i])] = c
		}
		if err := lambda(data, i); err != nil {
			return err
		}
		return nil
	})
}

func IterateCsvMapGroupErrors(r *csv.Reader, expectedHeaders []string, lambda MapIteratorGroupErrors) error {
	errs := &MultipleError{}
	if err := IterateCsvMap(r, expectedHeaders, func(row map[string]string, line int) error {
		lambda(row, line, errs)
		return nil
	}); err != nil {
		return err
	}
	return errs.Return()
}

// Checks for the expected headers and returns the
// column number for the last column required
func CheckCsvHeaders(row, expectedHeaders []string) (int, error) {
	headers := CsvHeadersToMap(row)
	var minRowCols int
	for _, x := range expectedHeaders {
		if v, ok := headers[x]; !ok {
			return -1, HeaderError(fmt.Sprintf("Missing header '%s'", x))
		} else if v > minRowCols {
			minRowCols = v
		}
	}
	return minRowCols, nil
}

func CsvHeadersToMap(row []string) map[string]int {
	m := make(map[string]int)
	for i, name := range row {
		m[CleanCell(name)] = i
	}
	return m
}

func CleanCell(str string) string {
	return zerowidth.RemoveZeroWidthCharacters(strings.TrimSpace(str))
}
