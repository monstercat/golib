package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON wraps a pointer to an object and expects the postgres data to be in
// either JSONB or JSON format. Implements sql.Scanner. If the provided value is
// nil, the provided item will not be touched.
type JSON struct {
	a any
}

// NewJSON creates a new JSON struct. Ensure that the provided parameter is a
// pointer.
func NewJSON(a any) *JSON {
	return &JSON{a}
}

// Scan implements sql.Scanner
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, bok := value.([]byte)
	s, sok := value.(string)
	if sok {
		bok = true
		b = []byte(s)
	}
	if !bok {
		return errors.New("Type assertion .([]byte) or .(string) failed.")
	}

	return json.Unmarshal(b, j.a)
}

// Value implements driver.Valuer
func (j *JSON) Value() (driver.Value, error) {
	return json.Marshal(j.a)
}
