package csv

import (
	"strings"
	"time"

	"github.com/monstercat/pgnull"
	"github.com/shopspring/decimal"

	"github.com/monstercat/golib/time"
)

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

func GetTimeFromRow(row map[string]string, field string, line int, required bool) (time.Time, error) {
	v, err := GetFieldFromRow(row, field, line, required)
	if err != nil {
		return time.Time{}, err
	}
	if v == "" {
		return time.Time{}, nil
	}
	xv, err := ParseCsvStringToTime(v)
	if err != nil {
		return time.Time{}, TransformCsvError(err, field, line)
	}
	return xv, nil
}

func GetNullTimeFromRow(row map[string]string, field string, line int) (pgnull.NullTime, error) {
	v, err := GetFieldFromRow(row, field, line, false)
	xv, err := ParseCsvStringToNullTime(v)
	if err != nil {
		return pgnull.NullTime{}, TransformCsvError(err, field, line)
	}
	return xv, nil
}

func ParseCsvStringToTime(str string) (time.Time, error) {
	xv, err := time.Parse(TimeFormat, str)
	if err != nil {
		// Check if they only provided date...
		// TODO refactor this to flex check method or something
		if yv, err2 := time.Parse(DateFormat, str); err2 == nil {
			xv = yv
		} else {
			return time.Time{}, &BadDateFormatError{Date: str}
		}
	}
	if !timeUtils.IsReasonableTime(xv) {
		return time.Time{}, &UnreasonableTimeCsvError{Date: str}
	}
	return xv, nil
}

func ParseCsvStringToNullTime(str string) (pgnull.NullTime, error) {
	if str == "" {
		return pgnull.NullTime{}, nil
	}
	date, err := ParseCsvStringToTime(str)
	if err != nil {
		return pgnull.NullTime{}, err
	}
	return pgnull.NullTime{date, true}, nil
}
