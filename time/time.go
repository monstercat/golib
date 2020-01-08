package timeUtils

import "time"

// Sometimes, NULL presents as a non-zero time.
// Consequently, checking IsZero is not enough.
func IsReasonableTime(t time.Time) bool {
	comp, _ := time.Parse(time.RFC3339, "1900-00-00T00:00:00Z")
	return t.After(comp)
}
