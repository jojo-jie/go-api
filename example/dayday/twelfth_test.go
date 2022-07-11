package dayday

import (
	"testing"
	"time"
)

func TestDay(t *testing.T) {
	nowTime := time.Now()
	t.Log(nowTime)
	t.Log(AddDay(nowTime, -2))

	t.Log(StartOfDay(nowTime, "Asia/Shanghai"))
	t.Log(EndOfDay(nowTime, "Asia/Shanghai"))

	location, _ := time.LoadLocation("Asia/Shanghai")
	first := time.Date(2022, time.July, 01, 12, 30, 20, 0, location)
	second := time.Date(2022, time.July, 29, 0, 0, 50, 0, location)
	t.Log("first and second are same: ", IsSameDay(first, second))
	t.Log("Days between start and end: ", DiffInDay(first, second))

	start := time.Now()
	end := time.Now().AddDate(0, 0, 26)
	days := FindNoOfDays(1, start, end)
	t.Log("No. of mondays between:", start, " and end:", end, "are: ", days)
}

func AddDay(t time.Time, days int) time.Time {
	newTime := t.AddDate(0, 0, days)
	return newTime
}

func StartOfDay(t time.Time, timezone string) time.Time {
	location, _ := time.LoadLocation(timezone)
	year, month, day := t.In(location).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func EndOfDay(t time.Time, timezone string) time.Time {
	location, _ := time.LoadLocation(timezone)
	year, month, day := t.In(location).Date()
	return time.Date(year, month, day, 23, 59, 59, 0, location)
}

func IsSameDay(first, second time.Time) bool {
	return first.YearDay() == second.YearDay() && first.Year() == second.Year()
}

// Find a number of days between given timestamps
func DiffInDay(start, end time.Time) int {
	// To find the minutes difference
	// end.Sub(start).Minutes()
	return int(end.Sub(start).Hours() / 24)
}

// Find a number of weekdays(Mon, Tue,…) between two dates

func FindNoOfDays(day int, start, end time.Time) int {
	totalDays := 0
	for start.Before(end) {
		if int(start.Weekday()) == day {
			totalDays++
		}
		start = AddDay(start, 1)
	}
	return totalDays
}

// Check if a given year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && year%100 != 0 || year%400 == 0
}


