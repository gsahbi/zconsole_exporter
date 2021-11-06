package util

import (
	"time"
)

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func Eod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 9999, t.Location())
}

func GetTodayRange() (string, string, error) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return "", "", err
	}
	now := time.Now()
	return Bod(now).In(utc).Format("2006-01-02T15:04:05.000Z07:00"),
		Eod(now).In(utc).Format("2006-01-02T15:04:05.000Z07:00"), nil
}
