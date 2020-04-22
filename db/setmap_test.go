package dbUtil

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/monstercat/pgnull"
)

func TestSetMap(t *testing.T) {

	var x = struct {
		Id            string `db:"id" setmap:"omitinsert"`
		Name          string
		CreatedAt     time.Time         `db:"created_at"`
		SomethingElse string            `db:"-"`
		StringArr     pq.StringArray    `db:"string_arr"`
		NullString    sql.NullString    `db:"null_string"`
		NullString2   pgnull.NullString `db:"null_string2"`
		NullInt1      pgnull.NullInt    `db:"null_int_1"`
	}{
		Id:            "1234566787",
		Name:          "Test Name",
		CreatedAt:     time.Now(),
		SomethingElse: "1234456",
		StringArr:     []string{"123", "456"},
		NullString:    sql.NullString{Valid: true, String: "test"},
		NullString2:   pgnull.NewNullString("test"),
		NullInt1:      pgnull.NewNullInt(23),
	}

	updateM := SetMap(&x, false)
	insertM := SetMap(&x, true)

	var tests = []struct {
		Result map[string]interface{}
		Key    string
		Exists bool
		Value  interface{}
	}{
		{
			Result: insertM,
			Key:    "id",
			Exists: false,
		},
		{
			Result: updateM,
			Key:    "id",
			Exists: true,
			Value:  x.Id,
		},
		{
			Result: insertM,
			Key:    "name",
			Exists: true,
			Value:  x.Name,
		},
		{
			Result: updateM,
			Key:    "name",
			Exists: true,
			Value:  x.Name,
		},
		{
			Result: insertM,
			Key:    "created_at",
			Exists: true,
			Value:  x.CreatedAt,
		},
		{
			Result: updateM,
			Key:    "created_at",
			Exists: true,
			Value:  x.CreatedAt,
		},
		{
			Result: insertM,
			Key:    "something_else",
			Exists: false,
		},
		{
			Result: updateM,
			Key:    "something_else",
			Exists: false,
		},
		{
			Result: insertM,
			Key:    "string_arr",
			Exists: true,
		},
		{
			Result: insertM,
			Key:    "null_string",
			Exists: true,
		},
		{
			Result: insertM,
			Key:    "null_string2",
			Exists: true,
		},
		{
			Result: insertM,
			Key:    "null_int_1",
			Exists: true,
		},
	}

	for _, tt := range tests {
		v, ok := tt.Result[tt.Key]
		if ok != tt.Exists {
			t.Errorf("Expected %s to %sexist but it does %sexist", tt.Key, boolToModifier(tt.Exists), boolToModifier(ok))
		}
		log.Print(tt.Key, v)
		if tt.Value != nil {
			if v != tt.Value {
				t.Errorf("Value %v expected. Got %v", tt.Value, v)
			}
		}
	}
}

func boolToModifier(b bool) string {
	if !b {
		return "not "
	}
	return ""
}
