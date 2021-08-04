package timeutil

import (
	"errors"
	"fmt"
	"time"
)

// Sometimes, NULL presents as a non-zero time.
// Consequently, checking IsZero is not enough.
func IsReasonableTime(t time.Time) bool {
	comp, _ := time.Parse(time.RFC3339, "1900-00-00T00:00:00Z")
	return t.After(comp)
}

func ParseTimeCheckNear(date, format string, target time.Time, leeway time.Duration) (bool, error) {
	found, err := time.Parse(format, date)
	if err != nil {
		return false, fmt.Errorf("error formatting %s to format %s, err %s", date, format, err)
	}

	if TimeNearTime(found, target, leeway) {
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("expected date %s +/- %s but got %s which is off by %f seconds", target, leeway, found, target.Sub(found).Seconds()))
}

func TimeNearTime(date time.Time, target time.Time, leeway time.Duration) bool {
	above := target.Add(leeway * -1)
	below := target.Add(leeway)
	return date.Before(below) && date.After(above)
}
