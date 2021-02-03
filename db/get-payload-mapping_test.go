package dbUtil

import (
	"testing"
)

func TestGetPayloadMapping(t *testing.T) {
	type dataModelType struct {
		FieldA bool   `db:"field_a" json:"fieldA"`
		FieldB string `db:"field_b" json:"fieldB" custom-map:"customFieldB"`
		FieldC string `db:"field_c"`
		FieldD bool   `db:"field_d" json:"fieldD" custom-map:"customFieldD"`
	}
	var dataModel dataModelType

	payload := map[string]interface{}{
		"customFieldB": "set field b",
		"fieldA":       false,
		"fieldC":       "another string",
		"fieldD":       true,
	}
	modified := GetPayloadMapping(&dataModel, payload)

	if modified["field_a"] != payload["fieldA"] {
		t.Errorf("Expecting field_a to be set to value of fieldA")
	}
	if modified["field_b"] != payload["customFieldB"] {
		t.Errorf("Expecting field_b to be set to value of customFieldB")
	}
	if modified["field_c"] != nil {
		t.Errorf("Expecting field_c not to be set without a json or custom-map tag")
	}
	if modified["field_d"] != nil {
		t.Errorf("Expecting field_d not to be set unless using the custom-map tag that's set")
	}

	payload = map[string]interface{}{
		"customFieldB": false,
		"fieldA":       "string",
	}
	modified = GetPayloadMapping(&dataModel, payload)
	if len(modified) != 0 {
		t.Errorf("Expecting no modified data when types do not match")
	}
}
