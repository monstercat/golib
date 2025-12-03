package daohelpers

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
)

// testIterator allows a fixed number of values to be returned.
type testIterator struct {
	idx    int
	values []string
}

func (i *testIterator) Next() (string, error) {
	defer func() {
		i.idx++
	}()

	if len(i.values) <= i.idx {
		return "", iterator.Done
	}

	return i.values[i.idx], nil
}

func (i *testIterator) Close() error {
	return nil
}

func TestMapIterator(t *testing.T) {
	ti := &testIterator{
		values: []string{"1", "2", "3"},
	}

	m := Map[string, int](ti, func(r string) int {
		i, err := strconv.Atoi(r)
		require.NoError(t, err)
		return i
	})

	i, err := m.Next()
	assert.Equal(t, 1, i)
	assert.Nil(t, err)

	i, err = m.Next()
	assert.Equal(t, 2, i)
	assert.Nil(t, err)

	i, err = m.Next()
	assert.Equal(t, 3, i)
	assert.Nil(t, err)

	_, err = m.Next()
	assert.ErrorIs(t, err, iterator.Done)
}
