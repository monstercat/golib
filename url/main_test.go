package url

import (
	"net/url"
	"testing"
)

func TestSortedQueryString(t *testing.T) {
	vals := url.Values{}
	vals.Add("aaa", "a test for you")
	vals.Add("xxx", "comma,separated")
	vals.Add("y:z", "last")
	vals.Add("aaa", "bbb")
	expected := "aaa=a+test+for+you&aaa=bbb&xxx=comma%2Cseparated&y%3Az=last"

	// We do many tests because Go randomizes the order it loops through maps, including
	// the url.Values
	for i := 0; i <= 200; i++ {
		if SortedQueryString(vals) != expected {
			t.Errorf("Failed at %d, \nExpected: '%s'\nFound:    '%s'", i, expected, SortedQueryString(vals))
		}
	}
}
