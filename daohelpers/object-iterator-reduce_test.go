package daohelpers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReduce(t *testing.T) {
	ti := &testIterator{
		values: []string{"1", "2", "3"},
	}
	r := Reduce[string, int](ti, func(agg int, item string) (int, bool) {
		i, err := strconv.Atoi(item)
		require.NoError(t, err)
		return agg + i, false
	})

	d, err := First[int](r)
	require.NoError(t, err)
	require.Equal(t, 6, d)
}

func TestMin(t *testing.T) {
	ti := &testIterator{
		values: []string{"1", "2", "3"},
	}
	m := Map[string, int](ti, func(r string) int {
		i, err := strconv.Atoi(r)
		require.NoError(t, err)
		return i
	})

	val, err := Min[int](m)
	require.NoError(t, err)
	require.Equal(t, 1, val)
}
