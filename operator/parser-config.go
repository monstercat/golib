package operator

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrParserInvalidMatch = errors.New("could not parser regexp matches")
)

// ParserRegexpDelegate is an interface that the Parser uses to help perform its work in producing operators. The Parser
// will delegate responsibility to the provided implementation.
type ParserRegexpDelegate interface {
	// Regexp function should return a single regular expression which parses for 1 operator, keeping in mind
	// remainders and modifiers.
	Regexp() (*regexp.Regexp, error)

	// ParseMatch parses the provided match to return:
	// - modifiers
	// - key
	// - value
	//
	// Depending on the return values, the Parser will either:
	// - exclude it
	// - include it as an operator
	// - include it as a remainder.
	ParseMatch(match []string) (mods []Modifier, key string, value []string, err error)

	// ParseMapKey decodes the input key by splitting out its modifiers.This is used for the Parser.ParseMap
	// functionality.
	ParseMapKey(inputKey string) (mods []Modifier, key string, err error)
}

// ParserDelegateSatDefaultModifiers is an optional interface allowing setting of "default" modifiers (e.g., in the case
// none are provided). This is for backwards compatibility and will be removed in the future.
//
// @DEPRECATED
type ParserDelegateSatDefaultModifiers interface {
	// SetDefaultModifiers should store the provided modifiers as the modifiers used in the REGEXP and the ParseMapKey
	// function IF another set of modifiers hasn't already been set.
	SetDefaultModifiers(m ...Modifier)
}

// ParserConfig implements ParserRegexpDelegate. It looks for operators of the form
//
//	[modifier-character][key][key-delimiter][str-start][stuff][str-end]
//
// For example, if StringStart and StringEnd were " and the keyDelimiter was =, an example operator would be
//
//	key="value"
//	!key="value"
//
// The key and the string delimiters are optional.
//
// Naming of ParserConfig is to ensure backwards compatibility for users that still use this. In the future, it will
// be changed to DefaultParserDelegate.
//
// @DEPRECATED
type ParserConfig struct {
	// String start and end characters.
	StringStart string
	StringEnd   string

	// Key delimiting character
	KeyDelimiter string

	// List of modifier runes.
	Modifiers []Modifier

	// Cache for the keyRegexp, so it only has to be generated once.
	keyRegexp *regexp.Regexp

	// Cache for the string regexp, so it only has to be generated once.
	stringRegexp *regexp.Regexp
}

func (c *ParserConfig) SetDefaultModifiers(m ...Modifier) {
	c.Modifiers = m
}

func (c *ParserConfig) regexpKeyString() string {
	return fmt.Sprintf("([%s]*)([\\w-]+)", regexp.QuoteMeta(string(c.Modifiers)))
}

func (c *ParserConfig) regexpString() string {
	return fmt.Sprintf("(%s%s)?([^%s%s\\s]+|%[3]s([^%[4]s]+)%[4]s)",
		c.regexpKeyString(),
		regexp.QuoteMeta(c.KeyDelimiter),
		regexp.QuoteMeta(c.StringStart),
		regexp.QuoteMeta(c.StringEnd),
	)
}

// Regexp function should return a single regular expression which parses for 1 operator, keeping in mind
// remainders and modifiers.
func (c *ParserConfig) Regexp() (r *regexp.Regexp, err error) {
	if c.stringRegexp == nil {
		c.stringRegexp, err = regexp.Compile(c.regexpString())
		if err != nil {
			return
		}
	}
	return c.stringRegexp, nil
}

// ParseMatch parses the provided match to return:
// - modifiers
// - key
// - value
//
// Match has the same values as regexp.FindSubmatchString
func (c *ParserConfig) ParseMatch(match []string) (mods []Modifier, key string, value []string, err error) {
	if len(c.Modifiers) == 0 {
		return c.parseMatchNoMods(match)
	}
	return c.parseMatch(match)
}

// parseMatch is meant to parse the match if there are modifiers.
func (c *ParserConfig) parseMatch(match []string) (mods []Modifier, key string, value []string, err error) {
	// Based on the regexp string, the indices are
	// 0 - the whole value  (ignore)
	// 1 - the key part including the delmiter & the tag (ignore)
	// 2 - the modifiers, if any
	// 3 - the key
	// 4 - the value but including the string start and string end
	// 5 - the value
	//
	// Thus, the length should be 6. If itt isn't we ignore.
	//
	// Based on the regexp, we don't need to check for Modifier validity as it is already valid out of the box.
	if len(match) != 6 {
		err = ErrParserInvalidMatch
		return
	}

	// Note that 4 and 5 cannot be ignored. There is a possibility that 5 is empty due to the way the regexp is
	// constructed.
	switch match[5] {
	case "":
		value = []string{match[4]}
	default:
		value = []string{match[5]}
	}

	key = match[3]
	for _, b := range []rune(match[2]) {
		mods = append(mods, Modifier(b))
	}
	return
}

// parseMatchNoMods is meant to parse the match if there are no modifiers.
func (c *ParserConfig) parseMatchNoMods(match []string) (mods []Modifier, key string, value []string, err error) {
	// Based on the regexp string, the indices are
	// 0 - the whole value  (ignore)
	// 1 - the key part including the delmiter & the tag (ignore)
	// 2 - the key
	// 3 - the value but including the string start and string end
	// 4 - the value
	//
	// Thus, the length should be 6. If itt isn't we ignore.
	//
	// Based on the regexp, we don't need to check for Modifier validity as it is already valid out of the box.
	if len(match) != 5 {
		err = ErrParserInvalidMatch
		return
	}

	// Note that 4 and 5 cannot be ignored. There is a possibility that 5 is empty due to the way the regexp is
	// constructed.
	switch match[4] {
	case "":
		value = []string{match[3]}
	default:
		value = []string{match[4]}
	}

	key = match[2]
	return
}

// ParseMapKey decodes the input key by splitting out its modifiers.This is used for the Parser.ParseMap
// functionality.
func (c *ParserConfig) ParseMapKey(inputKey string) (mods []Modifier, key string, err error) {
	if c.keyRegexp == nil {
		c.keyRegexp, err = regexp.Compile(c.regexpKeyString())
		if err != nil {
			return
		}
	}

	matches := c.keyRegexp.FindStringSubmatch(inputKey)

	// There should be 3 matches.
	// 0 - The whole key
	// 1 - The modifiers, if any
	// 2 - The actual key without modifiers.
	if len(matches) != 3 {
		err = ErrParserInvalidMatch
		return
	}

	key = matches[2]
	for _, b := range []rune(matches[1]) {
		mods = append(mods, Modifier(b))
	}
	return
}
