package util

import "time"

// Match the format  --->  "YYYY/MM/DD HH:mm:ss"
const EventDateLayout = "2006/01/02 15:04:05"

// Match the format -> "HH:mm"
const TimeLayout = "15:04"

func TimeParse(timeStr string) (*time.Time, error) {
	parsedTime, err := time.Parse(TimeLayout, timeStr)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

func EventDateParse(dateStr string) (*time.Time, error) {
	parsedTime, err := time.Parse(EventDateLayout, dateStr)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

func CountWeeksBetween(beginTime time.Time, endTime time.Time) int {
	duration := endTime.Sub(beginTime)
	weeks := int(duration.Hours() / (24 * 7))

	return weeks
}

// Return start of the week given time.Time (Sunday, 00:00)
func StartOfWeek(t time.Time) time.Time {
	// Calculate the number of days to subtract from the input time
	daysToSubtract := int(t.Weekday())

	// Subtract the days from the input time
	startOfWeek := t.AddDate(0, 0, -daysToSubtract)

	// Reset time to 00:00 (midnight)
	startOfWeekMidnight := time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	return startOfWeekMidnight
}
