package page

import (
	"strconv"
	"strings"
)

const (
	FieldPage   = "page"
	FieldLimit  = "limit"
	FieldOffset = "offset"
	FieldSort   = "sort"
)

// Count, Limit, Offset, Total are used for setting results of a query. The query performed may change any of these
// properties due to limitations or desired logic.
type Page struct {
	Fields map[string][]string

	// Count is used to determine the amount of results returned from the query.
	Count int

	// Limit is used to determine the total amount requested for the query.
	Limit int

	// Offset is used to determine where the cursor in the query should start.
	Offset int

	// Total is used to tell how many total results are available to query. A value of 0 or less can be considered as
	// unset.
	Total int
}

func (p *Page) Clone() *Page {
	return &Page{
		Fields: p.Fields,
		Offset: p.Offset,
		Limit:  p.Limit,
		Count:  p.Count,
		Total:  p.Total,
	}
}

func (p *Page) Get(key string) (string, bool) {
	xs, ok := p.Fields[key]
	if !ok || len(xs) == 0 {
		return "", false
	}
	return xs[0], true
}

func (p *Page) Set(key string, values ...string) *Page {
	if p.Fields == nil {
		p.Fields = make(map[string][]string)
	}
	p.Fields[key] = values
	return p
}

func (p *Page) SetBool(key string, values ...bool) *Page {
	var xs []string
	for _, val := range values {
		var str string
		if val {
			str = "true"
		} else {
			str = "false"
		}
		xs = append(xs, str)
	}
	return p.Set(key, xs...)
}

func (p *Page) SetInt(key string, values ...int) *Page {
	var xs []string
	for _, val := range values {
		xs = append(xs, strconv.Itoa(val))
	}
	return p.Set(key, xs...)
}

func (p *Page) SetCSV(key string, values ...string) *Page {
	return p.Set(key, strings.Join(values, ","))
}

func (p *Page) SetSort(sorts ...string) *Page {
	return p.SetCSV(FieldSort, sorts...)
}

func (p *Page) Add(key string, value ...string) *Page {
	if p.Fields == nil {
		p.Fields = make(map[string][]string)
	}
	p.Fields[key] = append(p.Fields[key], value...)
	return p
}

func (p *Page) Delete(key string) *Page {
	delete(p.Fields, key)
	return p
}

func (p *Page) Exists(key string) bool {
	_, ok := p.Fields[key]
	return ok
}

func (p *Page) GetBool(key string) (bool, bool) {
	if xs, ok := p.Fields[key]; !ok {
		return false, false
	} else if len(xs) == 0 {
		return false, true
	} else {
		return parseBoolStr(xs[0]), true
	}
}

func (p *Page) GetBooleans(key string) ([]bool, bool) {
	xs, ok := p.Fields[key]
	if !ok {
		return nil, false
	}
	var ys []bool
	for _, x := range xs {
		ys = append(ys, parseBoolStr(x))
	}
	return ys, true
}

func (p *Page) GetInt(key string) (int, bool) {
	if xs, ok := p.Fields[key]; !ok {
		return 0, false
	} else if len(xs) == 0 {
		return 0, true
	} else {
		i, _ := strconv.Atoi(xs[0])
		return i, true
	}
}

func (p *Page) GetIntegers(key string) ([]int, bool) {
	xs, ok := p.Fields[key]
	if !ok {
		return nil, false
	}
	var ys []int
	for _, x := range xs {
		i, _ := strconv.Atoi(x)
		ys = append(ys, i)
	}
	return ys, true
}

// This is special logic that is commonly used. The page number should be 1 to infinity in any instance thus we use this
// logic throughout applications.
func (p *Page) GetPageNumber() (int, bool) {
	i, ok := p.GetInt(FieldPage)
	if !ok {
		return 1, false
	} else if i < 1 {
		i = 1
	}
	return i, true
}

// Use this to take all CSV values of a key into a single array.
func (p *Page) GetCSV(key string) []string {
	xs, ok := p.Fields[key]
	if !ok {
		return nil
	}
	var arr []string
	for _, x := range xs {
		ys := strings.Split(x, ",")
		for _, y := range ys {
			arr = append(arr, strings.TrimSpace(y))
		}
	}
	return arr
}

func (p *Page) GetSorts() []string {
	return p.GetCSV(FieldSort)
}

func (p *Page) Str(key string) string {
	str, _ := p.Get(key)
	return str
}

func (p *Page) Int(key string) int {
	i, _ := p.GetInt(key)
	return i
}

func (p *Page) Bool(key string) bool {
	b, _ := p.GetBool(key)
	return b
}

func (p *Page) PageNumber() int {
	i, _ := p.GetPageNumber()
	return i
}

func (p *Page) Default(key, replacement string) string {
	val, ok := p.Get(key)
	if !ok {
		return replacement
	}
	return val
}

func (p *Page) DefaultBool(key string, replacement bool) bool {
	val, ok := p.GetBool(key)
	if !ok {
		return replacement
	}
	return val
}

func (p *Page) DefaultInt(key string, replacement int) int {
	val, ok := p.GetInt(key)
	if !ok {
		return replacement
	}
	return val
}

func (p *Page) ApplyPageNumber(num, limit int) *Page {
	if num <= 0 {
		num = 1
	}
	p.Offset = (num - 1) * limit
	p.Limit = limit
	return p
}

func (p *Page) ApplyDefault(key string, values ...string) *Page {
	if !p.Exists(key) {
		p.Add(key, values...)
	}
	return p
}

func (p *Page) CopyFrom(b *Page, keys ...string) *Page {
	for _, key := range keys {
		p.Fields[key] = b.Fields[key]
	}
	return p
}

func NewPage() *Page {
	return &Page{
		Fields: map[string][]string{},
		Total:  -1,
	}
}

func NewPageFromMap(m map[string][]string) *Page {
	p := &Page{
		Fields: m,
		Total:  -1,
	}
	p.Limit = p.Int(FieldLimit)
	p.Offset = p.Int(FieldOffset)
	return p
}

func parseBoolStr(str string) bool {
	check := strings.ToLower(str)
	switch check {
	case "yes", "true", "t", "y", "1":
		return true
	}
	return false
}
