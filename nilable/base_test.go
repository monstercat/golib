package nilable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase_IsNil(t *testing.T) {
	isNil := &Base{Nil: true}
	isNotNil := &Base{Nil: false}

	assert.True(t, isNil.IsNil())
	assert.False(t, isNotNil.IsNil())
}
