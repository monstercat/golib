package strutil

import "testing"

func TestIsEqual(t *testing.T) {
	a := []string{"1", "2", "3"}
	b := []string{"3", "2", "1"}
	c := []string{"1", "2", "3"}

	if IsEqual(a, b) {
		t.Error("a and b are not the same")
	}
	if !IsEqual(a, a) {
		t.Error("a is the same as a")
	}
	if !IsEqual(b, b) {
		t.Error("b is the same as b")
	}
	if !IsEqual(a, c) {
		t.Error("a is the same as c")
	}
}
