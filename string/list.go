package strutil

func StringInList(xs []string, a string) bool {
	return Index(xs, a) != -1
}

func Index(xs []string, a string) int {
	for i, g := range xs {
		if a == g {
			return i
		}
	}
	return -1

}

func IsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, aa := range a {
		if aa != b[i] {
			return false
		}
	}
	return true
}