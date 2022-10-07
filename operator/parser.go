package operator

import (
	"errors"
	"strings"
)

var (
	ErrInvalidDelegate = errors.New("invalid delegate")
)

// Parser parses maps/strings into operators using the provided delegate to modify its behaviour. Currently, the only
// delegate possible is the Regexp delegate. This delegate will ParseMatch based on a Regexp returned by the delegate
// and a corresponding ParseMatch function.
type Parser struct {
	// Delegate modifies the operation of the parser.
	//
	// TODO: Currently, only a Regexp parser is supported, but in the future, this should be extracted to extended to
	//   allow for other types of parsers.
	Delegate ParserRegexpDelegate
}

// NewParser generates a new parser based on the provided delegate. If defined, it will also set default modifiers.
//
// Unfortunately, the delegate does not work because we used to take in ParserConfig (note: missing the pointer).
// However, it currently wants &ParserConfig. Thus, we use interface{}.
func NewParser(delegate interface{}) (*Parser, error) {
	var d ParserRegexpDelegate
	switch v := delegate.(type) {
	case ParserRegexpDelegate:
		d = v
	case ParserConfig:
		d = &v
	default:
		return nil, ErrInvalidDelegate
	}

	if v, ok := d.(ParserDelegateSatDefaultModifiers); ok {
		v.SetDefaultModifiers(Modifiers...)
	}

	return &Parser{
		Delegate: d,
	}, nil
}

// ParseMap parses a map into operators. This can be used, for example, by passing the query string url.Values
// as operators. In this implementation, any query string parameters that do not have a key are ignored. Thus, no
// "remainders" will be returned.
func (p *Parser) ParseMap(m map[string][]string) Operators {
	operators := Operators{
		Values:     make(map[string][]Operator),
		Remainders: make([]Operator, 0, 10),
	}
	for k, vv := range m {
		// If no values, we ignore.
		if len(vv) == 0 {
			continue
		}

		// In the case that k is empty string, we can ignore.
		if len(k) == 0 {
			continue
		}

		mods, key, err := p.Delegate.ParseMapKey(k)
		if err != nil {
			operators.Errors = append(operators.Errors, &ParseMapKeyError{
				Base:     err,
				InputKey: k,
			})
			continue
		}
		operators.AddOperator(key, Operator{
			Values:    vv,
			Modifiers: mods,
		})
	}
	return operators
}

// Parse parses a string into operators based on the regexp generated from the parser config.
func (p *Parser) Parse(str string) Operators {
	str = strings.TrimSpace(str)

	operators := Operators{
		Values:     make(map[string][]Operator),
		Remainders: make([]Operator, 0, 10),
	}

	// Get the regexp. If we cannot, just return the operators.
	r, err := p.Delegate.Regexp()
	if err != nil {
		operators.Errors = append(operators.Errors, &ParseRegexpError{Base: err})
		return operators
	}

	// Get all matches
	matches := r.FindAllStringSubmatch(str, -1)
	if matches == nil {
		return operators
	}

	for _, m := range matches {
		mods, key, value, err := p.Delegate.ParseMatch(m)
		if err != nil {
			// On any error, it should continue. Note that this is for backwards compatibility. The function signature
			// does not allow for parser errors. In the case of parser errors, the whole match should be ignored.
			continue
		}

		// Ignore anything that is missing the value.
		if len(value) == 0 {
			continue
		}

		// if the key is empty, it's a remainder. We ignore the mods (there shouldn't be any).
		if key == "" {
			operators.AddRemainder(value...)
			continue
		}

		// Otherwise, we add the operator
		operators.AddOperator(key, Operator{
			Values:    value,
			Modifiers: mods,
		})
	}

	return operators
}
