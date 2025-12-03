package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT(t *testing.T) {
	assert.Equal(t, "field", T("", "field"))
	assert.Equal(t, "t.f", T("t", "f"))
	assert.Equal(t, "t.f", T("t.", "f"))
	assert.Equal(t, "f", T(".", "f"))
}
