package daohelpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	ti := &testIterator{
		values: []string{"1", "2", "3"},
	}

	filtered := Filter(ti, func(v string) bool {
		return v == "2"
	})

	collected, err := Collect[string](filtered)
	require.NoError(t, err)
	require.Equal(t, []string{"2"}, collected)
}
