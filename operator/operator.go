package operator

type Operators struct {
	// Map of key to operator
	Values     map[string][]Operator

	// Remainders
	Remainders []string
}

func (o *Operators) AddRemainder(str string) {
	if len(str) == 0 {
		return
	}
	o.Remainders = append(o.Remainders, str)
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

type Operator struct {
	Value     string
	Modifiers []Modifier
}

func (o *Operator) Has(m Modifier) bool {
	if o.Modifiers == nil || len(o.Modifiers) == 0 {
		return false
	}
	for _, mod := range Modifiers {
		if m == mod {
			return true
		}
	}
	return false
}
