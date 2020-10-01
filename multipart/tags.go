package multipart

import "strings"

type tags struct {
	Ignore    bool
	FieldName string
}

func (t *tags) Parse(str string) {
	p := strings.Split(str, ",")
	if len(p) == 0 {
		return
	}
	t.FieldName = p[0]
	if t.FieldName == "-" {
		t.Ignore = true
		return
	}
	if len(p) == 1 {
		return
	}
	for _, v := range p[1:] {
		switch strings.TrimSpace(v) {
		case "ignore", "-":
			t.Ignore = true
		}
	}
}

