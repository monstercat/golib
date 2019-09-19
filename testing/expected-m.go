package expected_m

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/monstercat/lm"
	"github.com/tidwall/gjson"
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
			return errors.New(fmt.Sprintf("For test `%s`, expect length: %s%d; got 0", string(comparator), k, test))
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
			"For test `%s`, expect length: %s%d; got %d\n",
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
		} else if k[len(k)-3:] == "(#)" {
			key := k[:len(k)-3]
			actualValue := result.Get(key).Value()
			if err := CheckGJSONLength(key, expectedValue, actualValue); err != nil {
				return err
			}
		} else if f, ok := expectedValue.(func(val interface{}) error); ok {
			return f(actualValue)
		} else {
			if !reflect.DeepEqual(actualValue, expectedValue) {
				msg := fmt.Sprintf("Unexpected JSON value at \"%v\". \n Expected: \"%v\" \n    Found: \"%v\"\n", k, expectedValue, actualValue)

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

func CheckDate(expectedStr string, format string) func(i interface{}) error {
	return func(json interface{}) error {
		expectedDate, err := time.Parse(format, expectedStr)
		if err != nil {
			return err
		}

		foundDate, err := time.Parse(format, json.(string))
		if err != nil {
			return err
		}

		expected := expectedDate.Format(format)
		found := foundDate.Format(format)

		if expected == found {
			return nil
		}

		return errors.New(fmt.Sprintf("Expected date %s but got %s", expected, found))
	}
}
