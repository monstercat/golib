package strutil

import "testing"

func TestCamelToSnakeCase(t *testing.T) {

	tests := map[string]string{
		"TagusId":               "tagus_id",
		"TryThisForGoodMeasure": "try_this_for_good_measure",
	}

	for k, v := range tests {
		out := CamelToSnakeCase(k)
		if out != v {
			t.Errorf("For %s, expected %s, got %s", k, v, out)
		}
	}
}
