package csv

import (
	"strings"
	"time"

	"github.com/monstercat/pgnull"
	"github.com/shopspring/decimal"
)

var defaultFormats = []string{
	time.RFC3339,
	TimeFormat,
	DateFormat,
}

const (
	TimeFormat = "2006-01-02 15:04:05 -0700"
	DateFormat = "2006-01-02"
)

func GetDecimalFromRow(row map[string]string, field string, line int, required bool) (decimal.Decimal, error) {
	v, err := GetFieldFromRow(row, field, line, required)
	if err != nil {
		return decimal.Decimal{}, err
	}
	v = strings.TrimSpace(v)
	d, err := decimal.NewFromString(v)
	if err != nil {
		csvErr := &BadDecimalFormatError{Value: v}
		csvErr.SetLine(line)
		csvErr.SetField(field)
		return decimal.Decimal{}, csvErr
	}
	return d, nil
}

func GetFieldFromRow(row map[string]string, field string, line int, required bool) (string, error) {
	v, ok := row[field]
	if !ok {
		if required {
			return "", &EmptyFieldError{
				Field: field,
				Line:  line,
			}
		} else {
			return "", nil
		}
	}
	return v, nil
}

func GetCustomTimeFromRow(fmts []string, row map[string]string, field string, line int, required bool) (time.Time, error) {
	v, err := GetFieldFromRow(row, field, line, required)
	if err != nil {
		return time.Time{}, err
	}
	if v == "" {
		return time.Time{}, nil
	}
	xv, err := ParseCsvStringToTime(fmts, v)
	if err != nil {
		return time.Time{}, TransformCsvError(err, field, line)
	}
	return xv, nil
}

func GetTimeFromRow(row map[string]string, field string, line int, required bool) (time.Time, error) {
	return GetCustomTimeFromRow(defaultFormats, row, field, line, required)
}

func GetCustomNullTimeFromRow(fmts []string, row map[string]string, field string, line int) (pgnull.NullTime, error) {
	return GetCustomNullTimeFromRowInLocation(fmts, row, field, line, time.UTC)
}

func GetCustomNullTimeFromRowInLocation(fmts []string, row map[string]string, field string, line int, loc *time.Location) (pgnull.NullTime, error) {
	v, err := GetFieldFromRow(row, field, line, false)
	xv, err := ParseCsvStringToNullTimeInLocation(fmts, v, loc)
	if err != nil {
		return pgnull.NullTime{}, TransformCsvError(err, field, line)
	}
	return xv, nil
}

func GetNullTimeFromRow(row map[string]string, field string, line int) (pgnull.NullTime, error) {
	return GetCustomNullTimeFromRow(defaultFormats, row, field, line)
}

func GetNullTimeFromRowInLocation(row map[string]string, field string, line int, loc *time.Location) (pgnull.NullTime, error) {
	return GetCustomNullTimeFromRowInLocation(defaultFormats, row, field, line, loc)
}

func ParseCsvStringToTimeInLocation(fmts []string, str string, loc *time.Location) (time.Time, error) {
	var xv time.Time
	var err error
	for _, fmt := range fmts {
		xv, err = time.ParseInLocation(fmt, str, loc)
		if err == nil {
			break
		}
	}
	if err != nil {
		return time.Time{}, &BadDateFormatError{Date: str}
	}
	return xv, nil
}

func ParseCsvStringToTime(fmts []string, str string) (time.Time, error) {
	return ParseCsvStringToTimeInLocation(fmts, str, time.UTC)
}

func ParseCsvStringToNullTime(fmts []string, str string) (pgnull.NullTime, error) {
	return ParseCsvStringToNullTimeInLocation(fmts, str, time.UTC)
}

func ParseCsvStringToNullTimeInLocation(fmts []string, str string, loc *time.Location) (pgnull.NullTime, error) {
	if str == "" {
		return pgnull.NullTime{}, nil
	}
	date, err := ParseCsvStringToTimeInLocation(fmts, str, loc)
	if err != nil {
		return pgnull.NullTime{}, err
	}
	return pgnull.NullTime{date, true}, nil
}
