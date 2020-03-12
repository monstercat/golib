package expectm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/monstercat/lm"
	"github.com/tidwall/gjson"

	mtime "github.com/monstercat/golib/time"
)

// This custom type is so you can't accidentally pass in the wrong
// map to a function that compares two mars
// EG: checkJSON(sentJSON, sentJSON) won't compile if the function
// expects ExpectedM as the 2nd parameter
type ExpectedM map[string]interface{}

func CheckJSONBytes(js []byte, expected *ExpectedM) error {
	res := gjson.ParseBytes(js)
	return CheckExpectedM(res, expected)
}

func CheckJSONString(js string, expected *ExpectedM) error {
	res := gjson.Parse(js)
	return CheckExpectedM(res, expected)
}

func CheckJSON(obj interface{}, expected *ExpectedM) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return CheckJSONBytes(bytes, expected)
}

func CheckGJSONLength(k string, expectedValue, actualValue interface{}) error {
	var test int
	var comparator rune

	if str, ok := expectedValue.(string); ok {
		var err error
		test, err = strconv.Atoi(str[1:])
		if err != nil {
			return err
		}
		comparator = rune(str[0])
	} else {
		v, err := lm.MustInt(expectedValue)
		if err != nil {
			return err
		}
		test = v
	}

	if actualValue == nil {
		var hasError bool
		if test > 0 {
			// if test is greater than zero, there is an error if comparator is not <
			hasError = comparator != '<'
		} else if test < 0 {
			hasError = comparator == '<'
		}
		if hasError {
			return errors.New(fmt.Sprintf("for test `%s`, expect length: %s%d; got 0", string(comparator), k, test))
		}
		return nil
	}

	actVal := int(actualValue.(float64))

	var hasError bool
	switch comparator {
	case '<':
		if test <= actVal {
			hasError = true
		}
	case '>':
		if test >= actVal {
			hasError = true
		}
	case '!':
		if test == actVal {
			hasError = true
		}
	default:
		if test != actVal {
			hasError = true
		}
	}

	if hasError {
		return errors.New(fmt.Sprintf(
			"for test `%s`, expect length: %s%d; got %d\n",
			k,
			string(comparator),
			test,
			actVal,
		))
	}

	return nil
}

func CheckExpectedM(result gjson.Result, expected *ExpectedM) error {
	for k, expectedValue := range *expected {
		actualValue := result.Get(k).Value()

		// Special for "field.#" when checking length of array that was returned
		if k[len(k)-1] == '#' {
			if err := CheckGJSONLength(k, expectedValue, actualValue); err != nil {
				return err
			}
			// Special for "field(#)" where field itself is a number value
			// EG "total(#)": ">30"
		} else if len(k) > 3 && k[len(k)-3:] == "(#)" {
			key := k[:len(k)-3]
			actualValue := result.Get(key).Value()
			if err := CheckGJSONLength(key, expectedValue, actualValue); err != nil {
				return err
			}
		} else if f, ok := expectedValue.(func(json interface{}) error); ok {
			if err := f(actualValue); err != nil {
				return errors.New(fmt.Sprintf("Error at \"%v\" %v", k, err))
			}
		} else {
			if !reflect.DeepEqual(actualValue, expectedValue) {
				msg := fmt.Sprintf("unexpected JSON value at \"%v\". \n Expected: \"%v\" \n    Found: \"%v\"\n JSON: %v", k, expectedValue, actualValue, result)

				// We don't really care about Go types for checking JSON
				// This lets us do {"numberField": 0} without having to do float64(0)
				if fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedValue) {
					continue
				}
				return errors.New(msg)
			}
		}

	}
	return nil
}

func CheckDate(expectedStr string, format string) func(json interface{}) error {
	return func(json interface{}) error {
		if json == nil {
			return errors.New(fmt.Sprintf("expected date %s but it was nil", expectedStr))
		}

		loc, _ := time.LoadLocation("Europe/London")

		expectedDate, err := time.Parse(format, expectedStr)
		if err != nil {
			return err
		}
		expectedDate = expectedDate.In(loc)

		foundDate, err := time.Parse(format, json.(string))
		if err != nil {
			return err
		}
		foundDate = foundDate.In(loc)

		expected := expectedDate.Format(format)
		found := foundDate.Format(format)

		if expected == found {
			return nil
		}

		return errors.New(fmt.Sprintf("expected date %s but got %s", expected, found))
	}
}

// Returns a handler function that can be used in an ExpectedM object to compare a date
// value, represented as a string such as in JSON, against an actual date passed in

// The leeway duration allows for the dates to be off by that much time
// This is useful when you are comparing one time to one that will happen after some code runs
// For example comparing a PostedDate of a blog post that you are creating versus the one
// actually stored in the database. Without the leeway then if they were off by milliseconds
// they would not be equal.

// Example:
// test := {
//   bodyShouldHave: ExpectedM{
//      "created": CheckDateClose(time.Now(), time.Second),
//   }
func CheckDateClose(target time.Time, leeway time.Duration) func(json interface{}) error {
	return func(val interface{}) error {
		if val == nil {
			return errors.New(fmt.Sprintf("expected date %s +/- %s", target, leeway))
		}

		valS := val.(string)
		format := "2006-01-02T15:04:05-07:00"
		_, err := mtime.ParseTimeCheckNear(valS, format, target, leeway)
		return err
	}
}
