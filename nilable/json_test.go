package nilable

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_MarshalJSON(t *testing.T) {
	p := New[time.Time]()
	p.SetNil(true)

	b, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	tn := time.Now()
	p.SetNil(false)
	p.SetValue(tn)

	tnStr := tn.Format(time.RFC3339Nano)
	b, err = json.Marshal(p)
	require.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("\"%s\"", tnStr), string(b))
}

func TestJSON_UnmarshalJSON(t *testing.T) {
	jsonStr := []byte("\"2023-03-34\"") // Invalid date

	// There's a problem with the JSON string, it will error.
	x := New[time.Time]()
	assert.Error(t, json.Unmarshal(jsonStr, x))

	// Sometimes, we don't want it to error.
	AddJSON(x, SetNilOnError[time.Time]())
	assert.NoError(t, json.Unmarshal(jsonStr, x))
	assert.True(t, x.IsNil())

	// Set a default for nil!
	AddJSON(
		x,
		SetNilOnError[time.Time](),
		DefaultForNil[time.Time](time.Time{}),
	)
	assert.NoError(t, json.Unmarshal(jsonStr, x))
	assert.False(t, x.IsNil())
	assert.Equal(t, time.Time{}, x.Value())
}

func TestJSON_UnmarshalJSON_InvalidTime(t *testing.T) {
	jsonStr := []byte("\"-0001-12-31T15:47:32-08:12\"")

	x := New[time.Time]()
	AddJSON(x, SetNilOnError[time.Time]())
	assert.NoError(t, json.Unmarshal(jsonStr, x))
	assert.True(t, x.IsNil())
}
