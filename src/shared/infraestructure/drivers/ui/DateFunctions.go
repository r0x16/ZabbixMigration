package ui

import "time"

func DateFormat(date time.Time, format string) string {
	switch format {
	case "human":
		format = "02 Jan 2006 15:04"
	case "datetime":
		format = time.DateTime
	case "date":
		format = time.DateOnly
	case "time":
		format = time.TimeOnly
	}
	return date.Format(format)
}
