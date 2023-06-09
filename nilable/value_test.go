package nilable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrimitive_SetValue(t *testing.T) {
	p := New[int]()

	p.SetValue(100)
	assert.Equal(t, 100, p.Value())

	p.SetNil(true)
	assert.True(t, p.IsNil())
	assert.Equal(t, 100, p.Value())
}
