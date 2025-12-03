package nilable

import "errors"

var (
	ErrNilNotAllowed = errors.New("nil not allowed")
)

// OptionParam are parameters provided to all Option implementations for their
// Execute method. It is not guaranteed that all items will be filled.
type OptionParam[T comparable] struct {
	// Value which the option should work on. This is always provided.
	Value *Value[T]

	// ByteDataProcessor is a processor that generates useful information
	// related to byte data.
	ByteDataProcessor OptionByteDataProcessor[T]

	// Err is any error that occurs during any process.
	Err error

	// Process is the name of the process that is currently being run.
	Process string
}

// OptionByteDataProcessor is an interface that provides data regarding the
type OptionByteDataProcessor[T comparable] interface {
	// ByteData returns the current byte data.
	ByteData() []byte

	// SetNil sets the underlying byte data to whatever represents Nil in the
	// encoding format. For example, for JSON, it would be "null".
	SetNil()

	// SetValue sets the value of the byte data processor to the value
	// provided.
	SetValue(t T) error
}

// Option is a generic interface for functionality modification of nullables.
type Option[T comparable] interface {
	// Execute the option.
	//
	// In the case that the option is inapplicable to the situation (e.g.,
	// the wrong process or invalid state of objects), it should return
	// nil
	//
	// An error result should stop the process of applying/executing options.
	// A nil result should allow other options to continue to be executed.
	//
	// At the end, the OptionParam.Err should be returned as an error by the
	// process. Therefore, Options can update this error object.
	Execute(p *OptionParam[T]) error
}

// OptionFn converts a function of a specific type into an Option.
type OptionFn[T comparable] func(p *OptionParam[T]) error

func (fn OptionFn[T]) Execute(p *OptionParam[T]) error {
	return fn(p)
}

// SetNilOnError is an option which causes the object to be set to nil in the
// case of an error. This should occur only during Marshaling and Unmarshaling.
func SetNilOnError[T comparable]() Option[T] {
	return OptionFn[T](func(p *OptionParam[T]) error {
		switch p.Process {
		case ProcessUnmarshal:
			if p.Err != nil {
				p.Value.SetNil(true)
				p.Err = nil
			}
		case ProcessMarshal:
			if p.Err != nil {
				p.Err = nil
				p.ByteDataProcessor.SetNil()
			}
		}
		return nil
	})
}

// DefaultForNil is an option which sets a default value in the case that the
// data to be marshalled or unmarshalled is nil.
func DefaultForNil[T comparable](def T) Option[T] {
	return OptionFn[T](func(p *OptionParam[T]) error {
		// If there Has been an error, do nothing.
		if p.Err != nil {
			return nil
		}

		// Process the nullable otherwise.
		switch p.Process {
		case ProcessUnmarshal:
			if p.Value.IsNil() {
				p.Value.SetNil(false)
				p.Value.SetData(def)
			}
		case ProcessMarshal:
			if !p.Value.IsNil() {
				return nil
			}
			if err := p.ByteDataProcessor.SetValue(def); err != nil {
				return err
			}
		}
		return nil
	})
}

func NoNil[T comparable]() Option[T] {
	return OptionFn[T](func(p *OptionParam[T]) error {
		// If there Has been an error, do nothing.
		if p.Err != nil {
			return nil
		}
		if p.Value.IsNil() {
			return ErrNilNotAllowed
		}
		return nil
	})
}

func EmptyStringAsNil() Option[string] {
	return OptionFn[string](func(p *OptionParam[string]) error {
		// If there Has been an error, do nothing.
		if p.Err != nil {
			return nil
		}
		// treat empty string as nil:
		if p.Value.Data() == "" {
			p.Value.SetNil(true)
		}
		return nil
	})
}

// ApplyOptions applies the options provided. The provided options are applied
// sequentially. All options are run until one returns a non-nil error, at which
// point the error will be returned to the calling function.
//
// In the case that all options are applied successfully, the error object in
// the provided OptionParam will be returned.
func ApplyOptions[T comparable](p *OptionParam[T], options []Option[T]) error {
	for _, o := range options {
		oErr := o.Execute(p)
		if oErr != nil {
			return oErr
		}
	}
	return p.Err
}

// Optionable is designed to be included in structs which require options. It
// satisfies the HasOptions interface.
type Optionable[T comparable] struct {
	Options []Option[T]
}

// AddOptions adds the provided set of options to the current set.
func (o *Optionable[T]) AddOptions(opts ...Option[T]) {
	o.Options = append(o.Options, opts...)
}
