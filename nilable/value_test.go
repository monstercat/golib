package nilable

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNilOrZero(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		expectedNil  bool
		expectedZero bool
	}{
		{name: "NilPointer", input: (*int)(nil), expectedNil: true, expectedZero: true},
		{name: "NonNilPointer", input: new(int), expectedNil: false, expectedZero: true},
		{name: "ZeroInt", input: 0, expectedNil: false, expectedZero: true},
		{name: "NonZeroInt", input: 1, expectedNil: false, expectedZero: false},
		{name: "ZeroString", input: "", expectedNil: false, expectedZero: true},
		{name: "NonZeroString", input: "hello", expectedNil: false, expectedZero: false},
		{name: "ZeroTime", input: time.Time{}, expectedNil: false, expectedZero: true},
		{name: "NonZeroTime", input: time.Now(), expectedNil: false, expectedZero: false},
		{name: "NilInterface", input: (interface{})(nil), expectedNil: true, expectedZero: false},
		{name: "ValidInterface", input: interface{}(0), expectedNil: false, expectedZero: true},
		{name: "NilSlice", input: []int(nil), expectedNil: true, expectedZero: true},
		{name: "NonNilSlice", input: []int{1}, expectedNil: false, expectedZero: false},
		{name: "NilMap", input: map[int]int(nil), expectedNil: true, expectedZero: true},
		{name: "NonNilMap", input: map[int]int{1: 2}, expectedNil: false, expectedZero: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNil, isZero := IsNilOrZero(tt.input)
			assert.Equal(t, tt.expectedNil, isNil)
			assert.Equal(t, tt.expectedZero, isZero)
		})
	}
}

func TestValue_IsNil(t *testing.T) {
	var x Value[int]
	x.SetNil(true)
	assert.True(t, x.IsNil())

	x.SetNil(false)
	assert.False(t, x.IsNil())
}

func TestValue_Data(t *testing.T) {
	var x Value[int]
	assert.Equal(t, 0, x.Data())
	x.SetData(100)
	assert.Equal(t, 100, x.Data())
}

func TestValue_JSON_Nil(t *testing.T) {
	var x Value[int]
	x.SetNil(true)

	b, err := json.Marshal(x)
	require.NoError(t, err)
	assert.Equal(t, "null", string(b))

	// Reset to zero value, and then unmarshal
	x.SetNil(false)
	require.NoError(t, json.Unmarshal(b, &x))
	assert.True(t, x.IsNil())
}

func TestValue_JSON_Int(t *testing.T) {
	var x Value[int]
	x.SetData(100)

	b, err := json.Marshal(x)
	require.NoError(t, err)
	assert.Equal(t, "100", string(b))

	// Reset to zero value, and then unmarshal
	x.SetData(0)
	require.NoError(t, json.Unmarshal(b, &x))
	assert.Equal(t, 100, x.Data())
}

func TestValue_JSON_SetNilForError(t *testing.T) {
	var x Value[time.Time]

	jsonStr := []byte("\"2023-03-34\"") // Invalid date

	// Normally, unmarshalling shoudl error.
	assert.Error(t, json.Unmarshal(jsonStr, x))

	// Add the option. We should get nil.
	x.AddOptions(SetNilOnError[time.Time]())
	assert.NoError(t, json.Unmarshal(jsonStr, &x))
	assert.True(t, x.IsNil())

	// Set a default for nil!
	x.AddOptions(DefaultForNil(time.Time{}))
	assert.NoError(t, json.Unmarshal(jsonStr, &x))
	assert.False(t, x.IsNil())
	assert.Equal(t, time.Time{}, x.Data())
}

func TestValue_JSON_InvalidTime(t *testing.T) {
	// This is a string that sometimes we receive from the f/e or send from
	// the backend.
	jsonStr := []byte("\"-0001-12-31T15:47:32-08:12\"")

	var x Value[time.Time]
	x.AddOptions(SetNilOnError[time.Time]())
	assert.NoError(t, json.Unmarshal(jsonStr, &x))
	assert.True(t, x.IsNil())
}

func TestValue_IsEqual(t *testing.T) {
	// Tests to ensure that equals works fine.
	assert.False(t, NewPositiveInt(1).IsEqual(NewPositiveInt(2)))
	assert.True(t, NewPositiveInt(2).IsEqual(NewPositiveInt(2)))
	assert.False(t, NewNonzeroTime(time.Now()).IsEqual(NewNonzeroTime(time.Now().Add(time.Second))))

	// This is to test that nanoseconds don't matter.
	assert.True(t, NewNonzeroTime(time.Now()).IsEqual(NewNonzeroTime(time.Now())))
	tm := time.Now()
	assert.True(t, NewNonzeroTime(tm).IsEqual(NewNonzeroTime(tm)))
}

func TestValue_Scan_String(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	var x Value[string]

	// Test scanning string.
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow("123"))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, "123", x.Data())

	// Test scanning string with []byte
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow([]byte("123")))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, "123", x.Data())

	// Test scanning string with nil
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(nil))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.True(t, x.IsNil())

	// Test returning error due to invalid type.
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(123))
	assert.Error(t, db.QueryRow("").Scan(&x))
}

func TestValue_Scan_Float(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	var x Value[float64]

	// Test scanning a float64
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(123.0))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, 123.0, x.Data())

	// Test scanning a float32
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(float32(123.0)))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, 123.0, x.Data())

	// Test scanning an int64
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(int64(123)))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, 123.0, x.Data())

	// Test scanning an int32
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(int32(123)))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, 123.0, x.Data())

	// Test scanning an int
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(int(123)))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, 123.0, x.Data())

	// Test scanning nil
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(nil))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.True(t, x.IsNil())

	// Test returning error due to invalid type.
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow("123"))
	assert.Error(t, db.QueryRow("").Scan(&x))
}

func TestValue_Scan_Time(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	var x Value[time.Time]
	tm := time.Now()

	// Test scanning a time
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(tm))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.Equal(t, tm, x.Data())

	// Test scanning nil
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(nil))
	require.NoError(t, db.QueryRow("").Scan(&x))
	assert.True(t, x.IsNil())

	// Test returning error due to invalid type.
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow("123"))
	assert.Error(t, db.QueryRow("").Scan(&x))
}

func BenchmarkValue_Scan_String(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < b.N; i++ {
		rows.AddRow("123")
	}

	// Mock setup
	mock.ExpectQuery("").
		WillReturnRows(rows)

	rr, err := db.Query("")
	require.NoError(b, err)

	var x Value[string]
	for i := 0; i < b.N; i++ {
		_ = rr.Scan(&x)
	}
}

func BenchmarkValue_Scan_StringAsByte(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < b.N; i++ {
		rows.AddRow([]byte("123"))
	}

	// Mock setup
	mock.ExpectQuery("").
		WillReturnRows(rows)

	rr, err := db.Query("")
	require.NoError(b, err)

	var x Value[string]
	for i := 0; i < b.N; i++ {
		_ = rr.Scan(&x)
	}
}

func BenchmarkValue_Scan_Float64(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < b.N; i++ {
		rows.AddRow(123.0)
	}

	// Mock setup
	mock.ExpectQuery("").
		WillReturnRows(rows)

	rr, err := db.Query("")
	require.NoError(b, err)

	var x Value[float64]
	for i := 0; i < b.N; i++ {
		_ = rr.Scan(&x)
	}
}

func BenchmarkValue_Scan_Float64AsInt(b *testing.B) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < b.N; i++ {
		rows.AddRow(int64(123))
	}

	// Mock setup
	mock.ExpectQuery("").
		WillReturnRows(rows)

	rr, err := db.Query("")
	require.NoError(b, err)

	var x Value[string]
	for i := 0; i < b.N; i++ {
		_ = rr.Scan(&x)
	}
}
