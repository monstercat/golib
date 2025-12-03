package postgres

import (
	"fmt"
	"strings"

	stringUtil "github.com/monstercat/golib/string"
)

// FilterFields will filter out fields from the provided fields list.
func FilterFields(fields []string, filterOut ...string) []string {
	if len(filterOut) == 0 {
		return fields
	}
	str := make([]string, 0, len(fields))
	for _, s := range fields {
		if stringUtil.StringInList(filterOut, s) {
			continue
		}
		str = append(str, s)
	}
	return str
}

// OnConflictUpdateSuffix creates a string that can be used in Suffix in the form of:
//
//	ON CONFLICT ON CONSTRAINT [constraint] SET x = excluded.x ....
func OnConflictUpdateSuffix(constraint string, fields ...string) string {
	onConflictFields := make([]string, 0, len(fields))
	for _, v := range fields {
		onConflictFields = append(onConflictFields, fmt.Sprintf("%s = excluded.%[1]s", v))
	}
	return "ON CONFLICT ON CONSTRAINT " + constraint + " DO UPDATE SET " + strings.Join(onConflictFields, ", ")
}

// OnConflictUpdateSuffixWithoutConstraint creates a string that can be used in Suffix in the form of:
//
//	ON CONFLICT [constraint] SET x = excluded.x ....
func OnConflictUpdateSuffixWithoutConstraint(conflicts string, fields ...string) string {
	onConflictFields := make([]string, 0, len(fields))
	for _, v := range fields {
		onConflictFields = append(onConflictFields, fmt.Sprintf("%s = excluded.%[1]s", v))
	}
	return "ON CONFLICT " + conflicts + " DO UPDATE SET " + strings.Join(onConflictFields, ", ")
}
