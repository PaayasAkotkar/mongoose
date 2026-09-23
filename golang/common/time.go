package common

import "time"

// FormatDateForClient This will return "On 2026-01-02 At 03:04:05 PM"
func FormatDateForClient(date time.Time) string {
	return date.Format("On 2006-01-02 At 03:04:05 PM")
}

// GenerateServerTime Generates: "03:45:01 PM"
func GenerateServerTime(time time.Time) string {
	return time.Format("03:04:05 PM")
}
