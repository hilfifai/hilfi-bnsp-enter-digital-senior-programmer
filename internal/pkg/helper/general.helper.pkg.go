package helper

import "time"

func CalculateFrequencyInMonths(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())

	totalMonths := years*12 + months + 1

	if totalMonths < 1 {
		return 1
	}

	return totalMonths
}

func ToNullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
