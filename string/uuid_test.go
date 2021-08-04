package strutil

import "testing"

func TestIsUuid(t *testing.T) {

	tests := []struct{
		Result bool
		Test string
	} {
		{true, "123e4567-e89b-12d3-a456-426655440000"},
		{false, "id:123e4567-e89b-12d3-a456-426655440000"},
		{false, "123e4567-e89g-12d3-a456-426655440000"},
		{true, "123ABCDE-E12F-a000-0033-159392aef039"},
		{false, "hi"},
		{false, "ag9320gnaobiz"},
		{false, "0123456789"},
	}

	for _, test := range tests {
		r := IsUuid(test.Test)
		if r != test.Result {
			t.Errorf("Expected %s to be %v, got %v", test.Test, test.Result, r)
		}
	}
}