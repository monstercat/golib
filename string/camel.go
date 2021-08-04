package strutil

import "unicode"

// Converts a string to Snake Case (lowercase)
func CamelToSnakeCase(in string) string {

	var str []rune
	for _, v := range in {
		if unicode.IsUpper(v) {
			str = append(str, '_', unicode.ToLower(v))
		} else {
			str = append(str, v)
		}
	}
	if len(str) > 0 && str[0] == '_' {
		str = str[1:]
	}
	return string(str)
}
