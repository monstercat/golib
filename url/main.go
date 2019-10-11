package url

import (
	"net/url"
	"sort"
)

func SortedUrlValues(vals url.Values) (sortedKeys []string, sortedKeyValues map[string][]string) {
	sortedKeys = make([]string, 0)
	sortedKeyValues = map[string][]string{}
	for key, _ := range vals {
		sortedKeys = append(sortedKeys, key)
		sortedKeyValues[key] = vals[key]
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})

	for _, key := range sortedKeys {
		unsorted := sortedKeyValues[key]
		sort.Slice(unsorted, func(i, j int) bool {
			return unsorted[i] < unsorted[j]
		})
	}

	return sortedKeys, sortedKeyValues
}

func SortedQueryString(vals url.Values) string {
	keys, keyVals := SortedUrlValues(vals)
	qs := ""
	first := true
	for _, key := range keys {
		for _, val := range keyVals[key] {
			if first {
				first = false
			} else {
				qs = qs + "&"
			}

			qs = qs + url.QueryEscape(key) + "=" + url.QueryEscape(val)
		}
	}

	return qs
}
