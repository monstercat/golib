package strutil

func StringInList(xs []string, a string) bool {
	for _, g := range xs {
		if a == g {
			return true
		}
	}
	return false
}

