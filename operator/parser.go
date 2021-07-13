package operator

import (
	"fmt"
	"regexp"
	"strings"

	stringutil "github.com/monstercat/golib/string"
)

var (
	ValidOperatorRegexp = regexp.MustCompile("^[A-z\\-]+$")
)

type Parser struct {
	// Modifiers that the parser will check.
	Modifiers      []Modifier
	KeyValueRegexp *regexp.Regexp
}

type ParserConfig struct {
	// String start and end characters.
	StringStart string
	StringEnd   string

	// Key delimiting character
	KeyDelimiter string
}

func (c *ParserConfig) regexpString() string {
	return fmt.Sprintf("(([\\w-]+)%s)?([^%s%s\\s]+|%[2]s([^%[3]s]+)%[3]s)",
		regexp.QuoteMeta(c.KeyDelimiter),
		regexp.QuoteMeta(c.StringStart),
		regexp.QuoteMeta(c.StringEnd),
	)
}

func (c *ParserConfig) Regexp() (*regexp.Regexp, error) {
	return regexp.Compile(c.regexpString())
}

func NewParser(config ParserConfig) (*Parser, error) {
	regexp, err := config.Regexp()
	if err != nil {
		return nil, err
	}
	return &Parser{
		Modifiers:      Modifiers,
		KeyValueRegexp: regexp,
	}, nil
}

func (p *Parser) ParseMap(m map[string][]string) Operators {
	operators := Operators{
		Values:     make(map[string][]Operator),
		Remainders: make([]Operator, 0, 10),
	}
	for k, vv := range m {
		if len(vv) == 0 {
			continue
		}
		currModifiers := make([]Modifier, 0, len(p.Modifiers))
		if mod := p.MatchModifier(k[0]); mod != nil {
			currModifiers = append(currModifiers, *mod)
			k = k[1:]
		}
		for _, v := range vv {
			operators.AddOperator(k, Operator{
				Value: v,
				Modifiers: currModifiers,
			})
		}
	}
	return operators
}

func (p *Parser) Parse(str string) Operators {
	str = strings.TrimSpace(str)

	operators := Operators{
		Values:     make(map[string][]Operator),
		Remainders: make([]Operator, 0, 10),
	}
	currModifiers := make([]Modifier, 0, len(p.Modifiers))

	for i := 0; i < len(str); i++ {
		// Ignore spaces
		if str[i] == ' ' {
			continue
		}

		// Check for existence of modifiers.
		if m := p.MatchModifier(str[i]); m != nil {
			currModifiers = append(currModifiers, *m)
			continue
		}

		// If no modifiers, we can continue to look for a key or a
		// value. If there is a ":" then it's a key. What comes next
		// is the value (unless there is a space, in which case we need to
		// ignore it).
		key, value, n := p.FindKeyOrValue(str[i:])
		if n == 0 {
			continue
		}

		if key == "" {
			operators.AddRemainder(value)
		} else {
			operators.AddOperator(key, Operator{
				Value:     value,
				Modifiers: currModifiers,
			})
		}

		// Reset the modifiers and increment the index
		i += n
		currModifiers = make([]Modifier, 0, len(p.Modifiers))
	}

	return operators
}

func (p *Parser) MatchModifier(str byte) *Modifier {
	for _, m := range p.Modifiers {
		if m.Matches(str) {
			return &m
		}
	}
	return nil
}

func (p *Parser) FindKeyOrValue(str string) (key string, value string, n int) {
	matchIdx := p.KeyValueRegexp.FindStringSubmatchIndex(str)

	// The match should occur right at the beginning
	if matchIdx == nil || matchIdx[0] != 0 || matchIdx[2] > 0 {
		return "", "", 0
	}

	// The length of the total match is the second argument.
	n = matchIdx[1]

	// After that there are either 2 or 4 subgroups.
	if matchIdx[2] == -1 || matchIdx[4] == -1 {
		key = ""
	} else {
		key = str[:matchIdx[5]]
	}

	// There's a problem! I need to have 10 indices
	if len(matchIdx) != 10 {
		return "", "", n
	}

	if matchIdx[8] == -1 {
		value = str[matchIdx[6]:matchIdx[7]]
	} else {
		value = str[matchIdx[8]:matchIdx[9]]
	}

	return
}

func (p *Parser) RemoveOperatorsFromString(str string, operators ...string) string {
	validOperators := make([]string, 0, len(operators))
	for _, o := range operators {
		if ValidOperatorRegexp.MatchString(o) {
			validOperators = append(validOperators, o)
		}
	}
	if len(validOperators) == 0 {
		return str
	}

	var res string
	for i := 0; i < len(str); i++ {
		l := len(res)

		// Ignore spaces
		if str[i] == ' ' {
			if l > 0 && res[l-1] != ' ' {
				res = res + " "
			}
			continue
		}

		search := i
		if p.MatchModifier(str[i]) != nil {
			search++
		}
		key, _, n := p.FindKeyOrValue(str[i:])
		if !stringutil.StringInList(validOperators, key) {
			res = res + str[i:search+n]
		}

		// as the for loop will continue to increment i we need to -1
		if search+n > i {
			i = search + n - 1
		}
		continue
	}
	return strings.TrimSpace(res)
}

func OperatorExtractComparator(v string) (string, string) {
	switch {
	case strings.HasPrefix(v, "="):
		return v[1:], "="
	case strings.HasPrefix(v, ">"), strings.HasPrefix(v, "<"):
		if v[1] == '=' {
			return v[2:], v[0:2]
		}
		return v[1:], v[0:1]
	default:
		return v, "="
	}
}
