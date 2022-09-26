package operator

type Operators struct {
	// Map of key to operator
	Values map[string][]Operator

	// Remainders
	Remainders []Operator

	// Any Errors related to parsing of the operator.
	Errors []error
}

// Equals returns true of the provided operators is the same as the current one.
func (o *Operators) Equals(b *Operators) bool {
	if len(o.Remainders) != len(b.Remainders) {
		return false
	}
	if len(o.Values) != len(b.Values) {
		return false
	}
	for _, r := range o.Remainders {
		var found bool
		for _, br := range b.Remainders {
			if r.Equals(br) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for k, v := range o.Values {
		existing, ok := b.Values[k]
		if !ok {
			return false
		}
		if len(existing) != len(v) {
			return false
		}
		for _, o := range v {
			var found bool
			for _, bo := range existing {
				if o.Equals(bo) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (o *Operators) AddRemainder(str ...string) {
	if len(str) == 0 {
		return
	}
	o.Remainders = append(o.Remainders, Operator{
		Values: str,
	})
}

func (o *Operators) AddOperator(key string, op Operator) {
	if len(op.Values) == 0 {
		return
	}
	m, ok := o.Values[key]
	if !ok {
		o.Values[key] = []Operator{op}
		return
	}
	o.Values[key] = append(m, op)
}

func (o *Operators) GetOperator(key string) []Operator {
	return o.Values[key]
}

func (o *Operators) GetOperatorValues(key string) []string {
	ops := o.GetOperator(key)
	return GetOperatorValues(ops)
}

// Operator defines a set of values and modifiers that go along with those values. Most operators contain only a single
// value, but prepare to handle multiple values.
//
// 2022-09-28 - Note that this is a breaking change from previous versions.
type Operator struct {
	// Values representative of this operator.
	Values []string

	// Modifiers associated with the operator.
	Modifiers []Modifier
}

// Equals returns true of the provided operator is the same as the current one.
func (o *Operator) Equals(b Operator) bool {
	if len(o.Modifiers) != len(b.Modifiers) {
		return false
	}
	if len(o.Values) != len(b.Values) {
		return false
	}
	for _, m := range o.Modifiers {
		var found bool
		for _, bm := range b.Modifiers {
			if bm == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, v := range o.Values {
		var found bool
		for _, bv := range b.Values {
			if v == bv {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (o *Operator) Has(m Modifier) bool {
	if o.Modifiers == nil || len(o.Modifiers) == 0 {
		return false
	}
	for _, mod := range o.Modifiers {
		if m == mod {
			return true
		}
	}
	return false
}

func GetOperatorValues(ops []Operator) []string {
	if len(ops) == 0 {
		return nil
	}
	xs := make([]string, 0, len(ops))
	for _, o := range ops {
		xs = append(xs, o.Values...)
	}
	return xs
}

func (o *Operators) Get(keys ...string) []Operator {
	if len(keys) == 0 {
		return nil
	}
	var os []Operator
	for _, k := range keys {
		op, ok := o.Values[k]
		if !ok {
			continue
		}
		os = append(os, op...)
	}
	return os
}
