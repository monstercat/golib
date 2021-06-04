package ezrouter

import (
	"net/http"
	"regexp"
	"strings"
)

type Route interface {
	Match(r *http.Request) *RouteMatches

	GetHandler() func(interface{})

	// Use the following to read the pattern or update it. For example if you want to prefix or suffix things.
	GetPattern() string
	SetPattern(p string) error
}

type RouteMatches struct {
	named  map[string]string
	values []string
}

func (m *RouteMatches) AllNamed() map[string]string {
	x := map[string]string{}
	for k, v := range m.named {
		x[k]=v
	}
	return x
}

func (m *RouteMatches) All() []string {
	return append([]string{}, m.values...)
}

func (m *RouteMatches) Add(name, value string) {
	if m.named == nil {
		m.named = make(map[string]string)
	}
	m.named[name] = value
	m.values = append(m.values, value)
}

func (m *RouteMatches) Named(name string) string {
	return m.named[name]
}

func (m *RouteMatches) Index(i int) string {
	if i >= len(m.values) {
		return ""
	}
	return m.values[i]
}

// Use this for exact matching routes with string comparison. Very basic.
type StringRoute struct {
	Method string
	Path   string
}

func (r *StringRoute) Match(req *http.Request) *RouteMatches {
	if req.Method != r.Method || req.URL.Path != r.Path {
		return nil
	}
	return &RouteMatches{}
}

func (r *StringRoute) GetPattern() string {
	return r.Path
}

func (r *StringRoute) SetPattern(p string) error {
	r.Path = p
	return nil
}

// You can use this to perform placeholder style routing such as "/article/:articleName". Valid placeholder patterns
// consist of characters A-z, 0-9, -, :, / and nothing else.
type PlaceholderRoute struct {
	// The HTTP method to filter on
	Method string

	// The placeholder pattern to match on e.g. /release/:releaseId
	Pattern string

	// The placeholders parsed from Pattern in order
	placeholders []string

	// The compiled regexp created form the Pattern field.
	re *regexp.Regexp
}

func (r *PlaceholderRoute) SetPattern(p string) error {
	r.Pattern = p
	return r.compile() // NOTICE maybe merge this into here.
}

func (r *PlaceholderRoute) GetPattern() string {
	return r.Pattern
}

// You should call this method before hand so that you don't run into silent Match errors.
func (r *PlaceholderRoute) compile() error {
	var err error
	re, err := regexp.Compile("(:[\\w\\d-_]+)")
	if err != nil {
		return err
	}
	placeholders := re.FindAllString(r.Pattern, -1)
	r.placeholders = []string{}
	str := r.Pattern
	for _, placeholder := range placeholders {
		str = strings.Replace(str, placeholder, "([\\w\\d-.]+)", 1)
		r.placeholders = append(r.placeholders, placeholder[1:])
	}

	r.re, err = regexp.Compile("^" + str + "$")
	return err
}

func (r *PlaceholderRoute) Match(req *http.Request) *RouteMatches {
	if req.Method != r.Method {
		return nil
	}

	// If you didn't compile this before hand we attempt to. On fail to compile it will fail silently.
	// Make sure you call Compile ahead of time.
	if r.re == nil {
		err := r.compile()
		if err != nil {
			return nil
		}
	}

	m := r.re.FindAllStringSubmatch(req.URL.Path, -1)
	if len(m) != 1 || len(m[0]) != len(r.placeholders)+1 {
		return nil
	}

	matches := &RouteMatches{}
	for i, mx := range m[0] {
		if i == 0 {
			continue
		}
		matches.Add(r.placeholders[i-1], mx)
	}

	return matches
}

type RegexpRoute struct {
	Regexp *regexp.Regexp
}

func (r *RegexpRoute) Match(req *http.Request) *RouteMatches {
	m := r.Regexp.FindAllStringSubmatch(req.URL.Path, -1)
	if len(m) == 0 {
		return nil
	}
	matches := &RouteMatches{}
	for _, m := range m[0] {
		matches.Add("", m)
	}
	return matches
}

func (r *RegexpRoute) GetPattern() string {
	return r.Regexp.String()
}

func (r *RegexpRoute) SetPattern(p string) error {
	var err error
	r.Regexp, err = regexp.Compile(p)
	return err
}
