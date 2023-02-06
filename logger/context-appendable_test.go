package logger

import "testing"

func TestAppendableContext(t *testing.T) {
	var l Logger = &Standard{}
	l = &Contextual{Logger: l, Context: map[string]interface{}{
		"A": 1,
		"B": 2,
	}}
	l = &Contextual{Logger: l, Context: map[string]interface{}{
		"C": 1,
		"D": 2,
	}}

	l.Log(SeverityInfo, NewContextualPayload("test log").Add(map[string]interface{}{
		"E": 3,
		"F": 4,
	}))
}

func TestAppendableContextNamed(t *testing.T) {
	var l Logger = &Standard{}
	l = &Contextual{Logger: l, Context: NewContext("ab", map[string]interface{}{
		"A": 1,
		"B": 2,
	})}
	l = &Contextual{Logger: l, Context: NewContext("cd", map[string]interface{}{
		"C": 1,
		"D": 2,
	})}
	l.Log(SeverityInfo, NewContextualPayload("test log").Add(map[string]interface{}{
		"E": 3,
		"F": 4,
	}))
}
