package helper

import (
	"os"
	"time"
)

func GetTimezone() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	return "Asia/Jakarta"
}

func GetDateStart(t time.Time) (time.Time, error) {
	timezone := GetTimezone()
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	timeInTZ := t.In(loc)
	dateStart := time.Date(
		timeInTZ.Year(),
		timeInTZ.Month(),
		timeInTZ.Day(),
		0, 0, 0, 0,
		loc,
	)

	return dateStart, nil
}
