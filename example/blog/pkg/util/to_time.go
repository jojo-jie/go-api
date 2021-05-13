package util

import "time"

var location *time.Location
var layout string

func init()  {
	location,_ = time.LoadLocation("Asia/Shanghai")
	layout = "2006-01-02 15:04:05"
}

func GetFormatTime(t time.Time) string {
	return t.In(location).Format(layout)
}

func GetCalculateTime(currentTimer time.Time, d string) (time.Time, error) {
	duration, err := time.ParseDuration(d)
	if err != nil {
		return time.Time{}, err
	}
	return currentTimer.Add(duration), nil
}
