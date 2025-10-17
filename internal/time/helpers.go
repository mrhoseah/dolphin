package time

import (
	"fmt"
	"html/template"
	"time"
)

// TemplateHelpers returns a map of template helper functions for time formatting
func TemplateHelpers() template.FuncMap {
	return template.FuncMap{
		"moment":        NewMoment,
		"now":           Now,
		"fromNow":       FromNow,
		"formatTime":    FormatTime,
		"formatDate":    FormatDate,
		"formatDateTime": FormatDateTime,
		"calendar":      Calendar,
		"humanize":      Humanize,
		"relativeTime":  RelativeTime,
		"isToday":       IsToday,
		"isYesterday":   IsYesterday,
		"isThisWeek":    IsThisWeek,
		"isThisYear":    IsThisYear,
	}
}

// FromNow is a helper function that takes a time.Time and returns a human-readable string
func FromNow(t time.Time) string {
	return NewMoment(t).FromNow()
}

// FormatTime is a helper function that formats a time.Time
func FormatTime(t time.Time, layout string) string {
	return NewMoment(t).Format(layout)
}

// FormatDate is a helper function that formats a time.Time as a date
func FormatDate(t time.Time) string {
	return NewMoment(t).FormatDate()
}

// FormatDateTime is a helper function that formats a time.Time as date and time
func FormatDateTime(t time.Time) string {
	return NewMoment(t).FormatDateTime()
}

// Calendar is a helper function that returns a calendar-style time string
func Calendar(t time.Time) string {
	return NewMoment(t).Calendar()
}

// Humanize is a helper function that returns a humanized time string
func Humanize(t time.Time) string {
	return NewMoment(t).Humanize()
}

// RelativeTime is a helper function that returns a relative time string
func RelativeTime(t time.Time) string {
	return NewMoment(t).RelativeTime()
}

// IsToday is a helper function that checks if a time is today
func IsToday(t time.Time) bool {
	return NewMoment(t).IsToday()
}

// IsYesterday is a helper function that checks if a time is yesterday
func IsYesterday(t time.Time) bool {
	return NewMoment(t).IsYesterday()
}

// IsThisWeek is a helper function that checks if a time is this week
func IsThisWeek(t time.Time) bool {
	return NewMoment(t).IsThisWeek()
}

// IsThisYear is a helper function that checks if a time is this year
func IsThisYear(t time.Time) bool {
	return NewMoment(t).IsThisYear()
}

// TimeAgo is a simple helper that returns "X time ago" format
func TimeAgo(t time.Time) string {
	return NewMoment(t).FromNow()
}

// TimeSince is a helper that returns time since a given time
func TimeSince(t time.Time) string {
	return NewMoment(t).FromNow()
}

// TimeUntil is a helper that returns time until a given time
func TimeUntil(t time.Time) string {
	now := time.Now()
	if t.After(now) {
		return NewMoment(t).FromNow()
	}
	return NewMoment(t).FromNow()
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := int(d.Hours() / 24)
	weeks := int(d.Hours() / (24 * 7))
	months := int(d.Hours() / (24 * 30))
	years := int(d.Hours() / (24 * 365))

	switch {
	case seconds < 60:
		if seconds <= 1 {
			return "a moment"
		}
		return fmt.Sprintf("%d seconds", seconds)
	case minutes < 60:
		if minutes == 1 {
			return "a minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	case hours < 24:
		if hours == 1 {
			return "an hour"
		}
		return fmt.Sprintf("%d hours", hours)
	case days < 7:
		if days == 1 {
			return "a day"
		}
		return fmt.Sprintf("%d days", days)
	case weeks < 4:
		if weeks == 1 {
			return "a week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	case months < 12:
		if months == 1 {
			return "a month"
		}
		return fmt.Sprintf("%d months", months)
	default:
		if years == 1 {
			return "a year"
		}
		return fmt.Sprintf("%d years", years)
	}
}

// FormatDurationAgo formats a duration as "X time ago"
func FormatDurationAgo(d time.Duration) string {
	return FormatDuration(d) + " ago"
}

// FormatDurationIn formats a duration as "in X time"
func FormatDurationIn(d time.Duration) string {
	return "in " + FormatDuration(d)
}
