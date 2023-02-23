package strutil

import "strings"

// SplitBefore includes the separator as the first character. This is similar
// to strings.SplitAfter but instead of splitting "after" the character, it does
// so before.
//
// Taken from https://go.dev/play/p/P4rZvBAuSih. Removed the first if statement
// to stay analogous to strings.SplitAfter
func SplitBefore(s, sep string) (out []string) {
	for len(s) > 0 {
		i := strings.Index(s[1:], sep)
		if i == -1 {
			out = append(out, s)
			break
		}

		out = append(out, s[:i+1])
		s = s[i+1:]
	}
	return out
}
