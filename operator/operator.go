package operator

type Operators struct {
	// Map of key to operator
	Values     map[string][]Operator

	// Remainders
	Remainders []Operator
}

func (o *Operators) AddRemainder(str string) {
	if len(str) == 0 {
		return
	}
	o.Remainders = append(o.Remainders, Operator{
		Value: str,
	})
}

func (o *Operators) AddOperator(key string, op Operator) {
	if len(op.Value) == 0 {
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

type Operator struct {
	Value     string
	Modifiers []Modifier
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
		xs = append(xs, o.Value)
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