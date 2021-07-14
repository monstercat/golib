package operator

import (
	"net/url"
	"regexp"
	"testing"
)

func TestParserConfig_Regexp(t *testing.T) {

	tests := []struct {
		c ParserConfig
		s string
	}{
		{
			c: ParserConfig{
				StringStart:  ">",
				StringEnd:    "<",
				KeyDelimiter: "-",
			},
			s: "(([\\w-]+)-)?([^><\\s]+|>[^<]+<)",
		},
		{
			c: ParserConfig{
				StringStart:  "\"",
				StringEnd:    "\"",
				KeyDelimiter: ":",
			},
			s: "(([\\w-]+):)?([^\"\"\\s]+|\"[^\"]+\")",
		},
		{
			c: ParserConfig{
				StringStart:  "?",
				StringEnd:    "!",
				KeyDelimiter: "+",
			},
			s: "(([\\w-]+)\\+)?([^\\?!\\s]+|\\?[^!]+!)",
		},
	}

	for i, test := range tests {
		s := test.c.regexpString()
		if s != test.s {
			t.Errorf("[%d] did not match", i)
		}
		_, err := regexp.Compile(s)
		if err != nil {
			t.Errorf("[%d] did not compile: %s", i, err)
		}
	}
}

func TestParser_RemoveOperatorsFromString(t *testing.T) {
	parser, err := NewParser(ParserConfig{
		StringStart:  "\"",
		StringEnd:    "\"",
		KeyDelimiter: ":",
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Remove []string
		Start  string
		Result string
	}{
		{
			Remove: []string{"state"},
			Start:  "state:12345",
			Result: "",
		},
		{
			Remove: []string{"state"},
			Start:  "state:12345 title:12345 43939",
			Result: "title:12345 43939",
		},
		{
			Remove: []string{"state"},
			Start:  "title:12345 state:12345 43939",
			Result: "title:12345 43939",
		},
		{
			Remove: []string{"state", "title"},
			Start:  "title:12345 state:12345 43939",
			Result: "43939",
		},
	}

	for _, test := range tests {
		res := parser.RemoveOperatorsFromString(test.Start, test.Remove...)
		if res != test.Result {
			t.Errorf("For `%s`, removing %#v... got `%s`, expected `%s`", test.Start, test.Remove, res, test.Result)
		}
	}
}

func TestParser_FindKeyOrValue(t *testing.T) {
	c := ParserConfig{
		StringStart:  "\"",
		StringEnd:    "\"",
		KeyDelimiter: ":",
	}
	regexp, err := c.Regexp()
	if err != nil {
		t.Fatal(err)
	}

	p := &Parser{
		KeyValueRegexp: regexp,
	}

	tests := []struct {
		s     string
		key   string
		value string
		n     int
	}{
		{
			s:     "tag:123456 tags:\"123454-fjgie\"",
			key:   "tag",
			value: "123456",
			n:     10,
		},
		{
			s:     "123456 tags:\"123454-fjgie\"",
			key:   "",
			value: "123456",
			n:     6,
		},
		{
			s:     "\"123454-fjgie\" tag:12309",
			key:   "",
			value: "123454-fjgie",
			n:     14,
		},
		{
			s:     "tags:\"123454-fjgie\" blah-blah:1234j",
			key:   "tags",
			value: "123454-fjgie",
			n:     19,
		},
	}

	for i, test := range tests {
		key, value, n := p.FindKeyOrValue(test.s)
		if n != test.n {
			t.Errorf("[%d] number of chars does not match, got %d; expect %d", i, n, test.n)
		}
		if key != test.key {
			t.Errorf("[%d] key does not match. Got %s expect %s", i, key, test.key)
		}
		if value != test.value {
			t.Errorf("[%d] value does not match. Got %s, expect %s", i, value, test.value)
		}
	}
}

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
						Value: "123456",
					}},
					"tags": {{
						Value: "123454-fjgie",
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
							Value: "123456",
						},
						{
							Value:     "54949",
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Value: "123454-fjgie",
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
							Value:     "54949",
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Value: "123454-fjgie",
					}},
				},
				Remainders: []Operator{
					{
						Value: "123456",
					},
				},
			},
		},
	}

	for i, test := range tests {
		ops := p.Parse(test.s)
		if len(ops.Remainders) != len(test.ops.Remainders) {
			t.Errorf("[%d] remainders not equal length. Got %d, want %d", i, len(ops.Remainders), len(test.ops.Remainders))
		}
		for _, r := range test.ops.Remainders {
			var found bool
			for _, rem := range ops.Remainders {
				if r.Value == rem.Value {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%d] expecting remainder %s, but cannot find in list", i, r.Value)
			}
		}
		for key, val := range ops.Values {
			testVal, ok := test.ops.Values[key]
			if !ok {
				t.Errorf("[%d] expecting an operator value at key %s; but there is none", i, key)
				continue
			}
			if len(val) != len(testVal) {
				t.Errorf("[%d] expecting %d operators for key %s; but got %d", i, len(testVal), key, len(val))
				continue
			}
			for j, v := range val {
				if v.Value != testVal[j].Value {
					t.Errorf("[%d] expecting operator value %s for key %s at idx %d; but got %s", i, testVal[j].Value, key, j, v.Value)
				}
				if len(v.Modifiers) != len(testVal[j].Modifiers) {
					t.Errorf("[%d] expecting %d modifiers but got %d for key %s at idx %d", i, len(testVal[j].Modifiers), len(v.Modifiers), key, j)
				}
				for k, m := range v.Modifiers {
					if m != testVal[j].Modifiers[k] {
						t.Errorf("[%d] Expecting modifier %v but got %v for key %s at idx %d", i, testVal[j].Modifiers[k], m, key, j)
					}
				}
			}
		}
	}
}

func TestParser_ParseMap(t *testing.T) {
	p, err := NewParser(ParserConfig{
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
				"tag": []string{"123456"},
				"tags": []string{"123454-fjgie"},
			},
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {{
						Value: "123456",
					}},
					"tags": {{
						Value: "123454-fjgie",
					}},
				},
			},
		},
		{
			s: url.Values{
				"tag": []string{"123456"},
				"tags": []string{"123454-fjgie"},
				"!tag": []string{"54949"},
			},
			ops: Operators{
				Values: map[string][]Operator{
					"tag": {
						{
							Value: "123456",
						},
						{
							Value:     "54949",
							Modifiers: []Modifier{ModifierNot},
						},
					},
					"tags": {{
						Value: "123454-fjgie",
					}},
				},
			},
		},
	}
	for i, test := range tests {
		ops := p.ParseMap(test.s)
		if len(ops.Remainders) != len(test.ops.Remainders) {
			t.Errorf("[%d] remainders not equal length. Got %d, want %d", i, len(ops.Remainders), len(test.ops.Remainders))
		}
		for _, r := range test.ops.Remainders {
			var found bool
			for _, rem := range ops.Remainders {
				if r.Value == rem.Value {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%d] expecting remainder %s, but cannot find in list", i, r.Value)
			}
		}
		for key, val := range ops.Values {
			testVal, ok := test.ops.Values[key]
			if !ok {
				t.Errorf("[%d] expecting an operator value at key %s; but there is none", i, key)
				continue
			}
			if len(val) != len(testVal) {
				t.Errorf("[%d] expecting %d operators for key %s; but got %d", i, len(testVal), key, len(val))
				continue
			}
			for j, v := range val {
				if v.Value != testVal[j].Value {
					t.Errorf("[%d] expecting operator value %s for key %s at idx %d; but got %s", i, testVal[j].Value, key, j, v.Value)
				}
				if len(v.Modifiers) != len(testVal[j].Modifiers) {
					t.Errorf("[%d] expecting %d modifiers but got %d for key %s at idx %d", i, len(testVal[j].Modifiers), len(v.Modifiers), key, j)
				}
				for k, m := range v.Modifiers {
					if m != testVal[j].Modifiers[k] {
						t.Errorf("[%d] Expecting modifier %v but got %v for key %s at idx %d", i, testVal[j].Modifiers[k], m, key, j)
					}
				}
			}
		}
	}
}