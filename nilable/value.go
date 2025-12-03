// Package nilable is a drop-in replacement for pgnull or sql.NullX packages
// specifically used by Label Manager. It includes a single struct [Value] which
// stores implements all interfaces required from the application logic to the
// data layer.
//
// Currently, the following are required:
// - JSON for application layer input/output
// - sql.Scanner for the data layer.
package nilable

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

const (
	// ProcessUnmarshal indicates for an option that the current process is an
	// unmarshalling process.
	ProcessUnmarshal = "unmarshal"

	// ProcessMarshal indicates for an option that the current process is a
	// marshalling process.
	ProcessMarshal = "marshal"
)

var (
	ErrInvalidType = errors.New("invalid type")
)

// IsNilOrZero checks if the given value is nil or has a zero value.
// It returns two booleans: isNil, and isZero.
// Note: Empty slices(nil slice) are not considered zero nor nil, use len() to
// check for empty slices.
func IsNilOrZero(v any) (bool, bool) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true, true
		}
		rv = rv.Elem()
	}

	// TODO: should we also attempt to check for reflect.Array?
	//  - IsValid returns true for both zero value and initialized values
	//  - IsNil panics
	//  - IsZero panics
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		isEmpty := rv.Len() == 0
		return isEmpty, isEmpty
	}

	if !rv.IsValid() {
		return true, false
	}

	return false, rv.IsZero()
}

// New creates a new value.
func New[T comparable]() *Value[T] {
	return &Value[T]{}
}

// NewValue initializes a new Value instance with the provided value and sets
// it as nil or non-nil based on its state.
// Zero values are considered not nil. Should zero values to be nil, please
// consider using explicit functions, such s NewPositiveInt, NewNonemptyString,
// or NewNonzeroTime.
func NewValue[T comparable](val T) *Value[T] {
	temp := New[T]()
	if isNil, _ := IsNilOrZero(val); isNil {
		temp.SetNil(true)
	} else {
		temp.SetData(val)
	}
	return temp
}

// NewPositiveInt return an int-like value which is nil if the provided value
// is less than 1.
func NewPositiveInt[T constraints.Integer](val T) *Value[T] {
	v := New[T]()
	if val > 0 {
		v.SetData(val)
	} else {
		v.SetNil(true)
	}
	return v
}

// NewNonemptyString returns a string-like value which is nil if empty.
func NewNonemptyString[T ~string](val T) *Value[T] {
	v := New[T]()
	if val != "" {
		v.SetData(val)
	} else {
		v.SetNil(true)
	}
	return v
}

// NewNonzeroTime returns a time value which is nil if zero.
func NewNonzeroTime(val time.Time) *Value[time.Time] {
	v := New[time.Time]()
	if !val.IsZero() {
		v.SetData(val)
	} else {
		v.SetNil(true)
	}
	return v
}

// Value stores a value that could be nil
type Value[T comparable] struct {
	// Allows options to be added to the value.
	Optionable[T]

	// Whether the value is nil
	notNil bool

	// The actual value.
	value T
}

// GetSetMapData implements IGetData from golib/struct-tag/iterator
func (v Value[T]) GetSetMapData() (any, bool) {
	if v.IsNil() {
		return nil, true
	}
	refVal := reflect.ValueOf(v.value)
	return v, refVal.IsZero()
}

func (v *Value[T]) IsEqual(c *Value[T]) bool {
	if !v.notNil && !c.notNil {
		return true
	}

	// Equality here does *not* work well with things like time.Time
	// because they cannot be compared that way. Pointers can be compared but
	// the comparison will not be what is expected by the user.
	switch cc := any(c.value).(type) {
	case time.Time:
		// For time.Time, we should just compare unix timestamps. Otherwise, it
		// will attempt to also compare the nanoseconds, which may not be stored
		// in the storage medium. We really don't care about the nanoseconds
		// any ways.
		return cc.Unix() == any(v.Data()).(time.Time).Unix()
	case equals[T]:
		return cc.Equal(v.Data())
	}

	return v.value == c.value
}

// IsNil returns true if the value is null
func (v *Value[T]) IsNil() bool {
	return !v.notNil
}

// Data returns the stored value without checking if it is Null. This may
// be a zero value or nil if IsNull is true. The reason it isn't called Value
// is due to sql.Scanner also using the function Value.
func (v *Value[T]) Data() T {
	return v.value
}

// SetNil sets whether the object is nil. It is self returning.
func (v *Value[T]) SetNil(b bool) *Value[T] {
	v.notNil = !b
	return v
}

// SetData sets the value, regardless of whether the value is null. It is
// self returning.
func (v *Value[T]) SetData(value T) *Value[T] {
	v.value = value
	v.notNil = true
	return v
}

// SetDataFromPointer sets the data from a pointer value. If the pointer
// value is nil, it will set nil to true. Otherwise, it copies in the data value
func (v *Value[T]) SetDataFromPointer(value *T) *Value[T] {
	if value == nil {
		v.SetNil(true)
	} else {
		v.SetData(*value)
	}
	return v
}

// UnmarshalJSON unmarshals a JSON object into the provided Nilable. The
// actual process is performed by the unmarshalJSON function. Options are simply
// applied here.
func (v *Value[T]) UnmarshalJSON(b []byte) error {
	err := v.unmarshalJSON(b)
	p := &OptionParam[T]{
		Err:               err,
		ByteDataProcessor: &JSONByteDataProcessor[T]{Data: b},
		Value:             v,
		Process:           ProcessUnmarshal,
	}
	return ApplyOptions[T](p, v.Options)
}

func (v *Value[T]) unmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		v.SetNil(true)
		return nil
	}

	// Unmarshal into the type provided.
	var t T
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	v.SetData(t)
	return nil
}

// MarshalJSON implements the Marshaller interface. The actual process is
// performed by the marshalJSON function. Options are simply applied here.
func (v Value[T]) MarshalJSON() ([]byte, error) {
	b, err := v.marshalJSON()
	p := &OptionParam[T]{
		Err:               err,
		ByteDataProcessor: &JSONByteDataProcessor[T]{Data: b},
		Value:             &v,
		Process:           ProcessMarshal,
	}
	if err := ApplyOptions[T](p, v.Options); err != nil {
		return nil, err
	}
	return p.ByteDataProcessor.ByteData(), p.Err
}

func (v *Value[T]) marshalJSON() ([]byte, error) {
	if v.IsNil() {
		return []byte("null"), nil
	}
	return json.Marshal(v.Data())
}

// JSONByteDataProcessor is a ByteDataProcessor
type JSONByteDataProcessor[T any] struct {
	Data []byte
}

// ByteData returns the current byte data.
func (p *JSONByteDataProcessor[T]) ByteData() []byte {
	return p.Data
}

// SetNil sets the underlying byte data to whatever represents Nil in the
// encoding format. For example, for JSON, it would be "null".
func (p *JSONByteDataProcessor[T]) SetNil() {
	p.Data = []byte("null")
}

// SetValue sets the value of the byte data processor to the value
// provided.
func (p *JSONByteDataProcessor[T]) SetValue(t T) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	p.Data = data
	return nil
}

// getStringForScan returns a string for scan.
func (v *Value[T]) getStringForScan(value interface{}) (string, error) {
	tm, valid := value.(string)
	if valid {
		return tm, nil
	}
	tm2, valid := value.([]byte)
	if valid {
		return string(tm2), nil
	}
	return "", ErrInvalidType
}

// StringScanner is a generic type which allows the scanning of a string.
// Return ErrInvalidType if error occurs.
type StringScanner interface {
	ScanString(str string) error
}

// Scan implements the Scanner interface.
func (v *Value[T]) Scan(value interface{}) error {
	if value == nil {
		v.SetNil(true)
		return nil
	}

	tm, valid := value.(T)
	if valid {
		v.SetData(tm)
		return nil
	}

	// Special scanning utilities.
	switch x := any(&v.value).(type) {
	case StringScanner:
		tm, err := v.getStringForScan(value)
		if err != nil {
			return err
		}
		if err := x.ScanString(tm); err != nil {
			return err
		}
		v.notNil = true
		return nil
	}

	// There isn't actually any perceivable penalty to doing this. The penalty
	// comes from converting []byte to string, but otherwise, the conversions
	// for example, for float, or the switch statement, do not slow down the
	// process.
	switch any(v.value).(type) {
	case uint64:
		switch val := value.(type) {
		case uint32:
			v.SetData(any(uint64(val)).(T))
		case uint:
			v.SetData(any(uint64(val)).(T))
		case int64:
			v.SetData(any(uint64(val)).(T))
		case int32:
			v.SetData(any(uint64(val)).(T))
		default:
			return ErrInvalidType
		}
		return nil
	case string:
		// In the case that the desired type is a string, we need to check
		// for []byte.
		tm, valid := value.([]byte)
		if !valid {
			return ErrInvalidType
		}

		// We need to do a double cast here. This is safe because we *know* that
		// []byte is castable to string, and we know that T == string
		v.SetData(any(string(tm)).(T))
		return nil
	case float64:
		switch val := value.(type) {
		case float32:
			v.SetData(any(float64(val)).(T))
		case int64:
			v.SetData(any(float64(val)).(T))
		case int32:
			v.SetData(any(float64(val)).(T))
		case int:
			v.SetData(any(float64(val)).(T))
		default:
			return ErrInvalidType
		}
		return nil
	default:
		return ErrInvalidType
	}
}

// Value implements the driver Valuer interface.
func (v Value[T]) Value() (driver.Value, error) {
	if v.IsNil() {
		return nil, nil
	}
	if v, ok := any(v.value).(driver.Valuer); ok {
		return v.Value()
	}
	switch xx := any(v.value).(type) {
	case uint64:
		return int64(xx), nil
	}
	return v.value, nil
}
