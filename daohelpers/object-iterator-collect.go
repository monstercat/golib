package daohelpers

import (
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"

	"github.com/monstercat/golib/dao/daohelpers"
)

// Collect gathers all elements from the given ObjectIterator into a slice and
// returns it along with any encountered error.
func Collect[R any](iter daohelpers.ObjectIterator[R]) ([]R, error) {
	rr := make([]R, 0, 50)
	for {
		r, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return rr, nil
		}
		if err != nil {
			return nil, err
		}

		rr = append(rr, r)
	}
}

// First retrieves the first element from the given ObjectIterator or returns
// an error if one occurs during iteration.
func First[R any](iter daohelpers.ObjectIterator[R]) (R, error) {
	r, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return r, nil
	}
	if err != nil {
		return r, err
	}
	return r, nil
}
