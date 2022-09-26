package operator

import (
	"net/url"
	"testing"
)

func TestParser_Parse(t *testing.T) {

	p, err := NewParser(ParserConfig{
		StringStart:  "\"",
		StringEnd:    "\"",
		KeyDelimiter: ":",
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		s   string
		ops Operators
	}{
		{
			s: " tag:123456 tags:\"123454-fjgie\" ",
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {{
						Values: []string{"123456"},
					}},
					"tags": {{
						Values: []string{"123454-fjgie"},
					}},
				},
			},
		},
		{
			s: " tag:123456 tags:\"123454-fjgie\" !tag:54949",
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {
						{
							Values: []string{"123456"},
						},
						{
							Values:    []string{"54949"},
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Values: []string{"123454-fjgie"},
					}},
				},
			},
		},
		{
			s: " 123456 tags:\"123454-fjgie\" !tag:54949",
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {
						{
							Values:    []string{"54949"},
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Values: []string{"123454-fjgie"},
					}},
				},
				Remainders: []Operator{
					{
						Values: []string{"123456"},
					},
				},
			},
		},
	}

	for i, test := range tests {
		ops := p.Parse(test.s)
		if !ops.Equals(&test.ops) {
			t.Errorf(`[%d] Operator is not as expected. 
       Expected: %#v
	   Actual: %#v
`,
				i,
				test.ops,
				ops,
			)
		}
	}
}

func TestParser_ParseMap(t *testing.T) {
	p, err := NewParser(&ParserConfig{
		StringStart:  "\"",
		StringEnd:    "\"",
		KeyDelimiter: ":",
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		s   url.Values
		ops Operators
	}{
		{
			s: url.Values{
				"tag":  []string{"123456"},
				"tags": []string{"123454-fjgie"},
			},
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {{
						Values: []string{"123456"},
					}},
					"tags": {{
						Values: []string{"123454-fjgie"},
					}},
				},
			},
		},
		{
			s: url.Values{
				"tag":  []string{"123456"},
				"tags": []string{"123454-fjgie"},
				"!tag": []string{"54949"},
			},
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {
						{
							Values: []string{"123456"},
						},
						{
							Values:    []string{"54949"},
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Values: []string{"123454-fjgie"},
					}},
				},
			},
		},
	}
	for i, test := range tests {
		ops := p.ParseMap(test.s)
		if !ops.Equals(&test.ops) {
			t.Errorf(`[%d] Operator is not as expected. 
       Expected: %#v
	   Actual: %#v
`,
				i,
				test.ops,
				ops,
			)
		}
	}
}
