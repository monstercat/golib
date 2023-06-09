package nilable

import "encoding/json"

// NewJSON wraps a nullable with the JSON version.
func NewJSON[T any](n Nilable[T]) *JSON[T] {
	// In the case it is already JSON wrapped, just ignore and
	// return original.
	if v, ok := n.(*JSON[T]); ok {
		return v
	}
	return &JSON[T]{
		Nilable: n,
	}
}

// JSON is a wrapper which adds JSON bindings to a nullable. It implements both
// the Nilable interface and the json.Marshaller/json.Unmarshaller interfaces.
type JSON[T any] struct {
	Nilable[T]
	Optionable[T]
}

// Unwrap returns the Nilable underneath.
func (j *JSON[T]) Unwrap() Nilable[T] {
	return j.Nilable
}

// UnmarshalJSON unmarshals a JSON object into the provided Nilable. The
// actual process is performed by the unmarshalJSON function. Options are simply
// applied here.
func (j *JSON[T]) UnmarshalJSON(b []byte) error {
	err := j.unmarshalJSON(b)
	p := &OptionParam[T]{
		Err:               err,
		ByteDataProcessor: &JSONByteDataProcessor[T]{Data: b},
		Nullable:          j.Nilable,
		Process:           ProcessUnmarshal,
	}
	return ApplyOptions[T](p, j.Options)
}

func (j *JSON[T]) unmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		j.SetNil(true)
		return nil
	}

	// Unmarshal into the type provided.
	var t T
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	j.SetValue(t)
	return nil
}

// MarshalJSON implements the Marshaller interface. The actual process is
// performed by the marshalJSON function. Options are simply applied here.
func (j *JSON[T]) MarshalJSON() ([]byte, error) {
	b, err := j.marshalJSON()
	p := &OptionParam[T]{
		Err:               err,
		ByteDataProcessor: &JSONByteDataProcessor[T]{Data: b},
		Nullable:          j.Nilable,
		Process:           ProcessMarshal,
	}
	if err := ApplyOptions[T](p, j.Options); err != nil {
		return nil, err
	}
	return p.ByteDataProcessor.ByteData(), p.Err
}

func (j *JSON[T]) marshalJSON() ([]byte, error) {
	if j.IsNil() {
		return []byte("null"), nil
	}
	return json.Marshal(j.Value())
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
