package stringutil

import "regexp"

var (
	UuidRegexp = regexp.MustCompile("^[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$")
)

func IsUuid(s string) bool {
	return UuidRegexp.Match([]byte(s))
}
