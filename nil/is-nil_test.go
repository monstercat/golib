package nil

import "testing"

type A interface {
	B()
}
type B struct{}

func (b *B) B() {}

func TestIsNil(t *testing.T) {
	// Initialize a nil variable
	var b *B

	test := func(a A) {
		if a == nil {
			t.Error("A should not be 'nil'")
		}
		if !IsNil(a) {
			t.Error("A should be nil.")
		}
	}
	test(b)

	// Test nil with other types of variables
	var i *int
	var f *float32
	var a []int
	var m map[string]interface{}

	x := []interface{} {i, f, a, m }
	for _, xx := range x {
		if xx == nil {
			t.Error("A should not be 'nil'")
		}
		if !IsNil(xx) {
			t.Error("A should be nil.")
		}
	}
}
