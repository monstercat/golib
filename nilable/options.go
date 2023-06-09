package nilable

// IsOptionable describes a Nilable which uses options.
type IsOptionable[T any] interface {
	// ReplaceOptions replaces the current set of options with the provided one.
	ReplaceOptions(opts ...Option[T])

	// AddOptions adds the provided set of options to the current set.
	AddOptions(opts ...Option[T])
}

// OptionByteDataProcessor is an interface that provides data regarding the
type OptionByteDataProcessor[T any] interface {
	// ByteData returns the current byte data.
	ByteData() []byte

	// SetNil sets the underlying byte data to whatever represents Nil in the
	// encoding format. For example, for JSON, it would be "null".
	SetNil()

	// SetValue sets the value of the byte data processor to the value
	// provided.
	SetValue(t T) error
}

// OptionParam are parameters provided to all Option implementations for their
// Execute method. It is not guaranteed that all items will be filled.
type OptionParam[T any] struct {
	// Nullable which the option should work on. This is always provided.
	Nullable Nilable[T]

	// ByteDataProcessor is a processor that generates useful information
	// related to byte data.
	ByteDataProcessor OptionByteDataProcessor[T]

	// Err is any error that occurs during any process.
	Err error

	// Process is the name of the process that is currently being run.
	Process string
}

// Option is a generic interface for functionality modification of nullables.
type Option[T any] interface {
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
type OptionFn[T any] func(p *OptionParam[T]) error

func (fn OptionFn[T]) Execute(p *OptionParam[T]) error {
	return fn(p)
}

// SetNilOnError is an option which causes the object to be set to nil in the
// case of an error. This should occur only during Marshaling and Unmarshaling.
func SetNilOnError[T any]() Option[T] {
	return OptionFn[T](func(p *OptionParam[T]) error {
		switch p.Process {
		case ProcessUnmarshal:
			if p.Err != nil {
				p.Nullable.SetNil(true)
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
func DefaultForNil[T any](def T) Option[T] {
	return OptionFn[T](func(p *OptionParam[T]) error {
		// If there Has been an error, do nothing.
		if p.Err != nil {
			return nil
		}

		// Process the nullable otherwise.
		switch p.Process {
		case ProcessUnmarshal:
			if p.Nullable.IsNil() {
				p.Nullable.SetNil(false)
				p.Nullable.SetValue(def)
			}
		case ProcessMarshal:
			if !p.Nullable.IsNil() {
				return nil
			}
			if err := p.ByteDataProcessor.SetValue(def); err != nil {
				return err
			}
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
func ApplyOptions[T any](p *OptionParam[T], options []Option[T]) error {
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
type Optionable[T any] struct {
	Options []Option[T]
}

// ReplaceOptions replaces the current set of options with the provided one.
func (o *Optionable[T]) ReplaceOptions(opts ...Option[T]) {
	o.Options = opts
}

// AddOptions adds the provided set of options to the current set.
func (o *Optionable[T]) AddOptions(opts ...Option[T]) {
	o.Options = append(o.Options, opts...)
}

// ReplaceOptions replaces the options of a nilable (if the nilable
// IsOptionable)
func ReplaceOptions[T any](n Nilable[T], opts ...Option[T]) {
	if o, ok := n.(IsOptionable[T]); ok {
		o.ReplaceOptions(opts...)
	}
}

// AddOptions adds to the options of a nilable (if the nilable
// IsOptionable).
func AddOptions[T any](n Nilable[T], opts ...Option[T]) {
	if o, ok := n.(IsOptionable[T]); ok {
		o.AddOptions(opts...)
	}
}
