package time

import (
	"fmt"
	"time"
)

// Moment provides human-readable time formatting similar to moment.js
type Moment struct {
	time time.Time
}

// NewMoment creates a new Moment instance from a time.Time
func NewMoment(t time.Time) *Moment {
	return &Moment{time: t}
}

// Now creates a Moment instance for the current time
func Now() *Moment {
	return &Moment{time: time.Now()}
}

// FromUnix creates a Moment instance from a Unix timestamp
func FromUnix(timestamp int64) *Moment {
	return &Moment{time: time.Unix(timestamp, 0)}
}

// FromString creates a Moment instance from a time string
func FromString(timeStr string) (*Moment, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}
	return &Moment{time: t}, nil
}

// Time returns the underlying time.Time
func (m *Moment) Time() time.Time {
	return m.time
}

// Format returns a formatted time string
func (m *Moment) Format(layout string) string {
	return m.time.Format(layout)
}

// FormatDate returns a formatted date string (YYYY-MM-DD)
func (m *Moment) FormatDate() string {
	return m.time.Format("2006-01-02")
}

// FormatTime returns a formatted time string (HH:MM:SS)
func (m *Moment) FormatTime() string {
	return m.time.Format("15:04:05")
}

// FormatDateTime returns a formatted date and time string
func (m *Moment) FormatDateTime() string {
	return m.time.Format("2006-01-02 15:04:05")
}

// FormatISO returns an ISO 8601 formatted string
func (m *Moment) FormatISO() string {
	return m.time.Format(time.RFC3339)
}

// FromNow returns a human-readable string describing the time relative to now
func (m *Moment) FromNow() string {
	now := time.Now()
	diff := now.Sub(m.time)

	// Handle future times
	if diff < 0 {
		diff = -diff
		return m.formatFuture(diff)
	}

	return m.formatPast(diff)
}

// formatPast formats a past time difference
func (m *Moment) formatPast(diff time.Duration) string {
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := int(diff.Hours() / (24 * 7))
	months := int(diff.Hours() / (24 * 30))
	years := int(diff.Hours() / (24 * 365))

	switch {
	case seconds < 60:
		if seconds <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes < 60:
		if minutes == 1 {
			return "a minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "an hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 7:
		if days == 1 {
			return "a day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case weeks < 4:
		if weeks == 1 {
			return "a week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case months < 12:
		if months == 1 {
			return "a month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "a year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// formatFuture formats a future time difference
func (m *Moment) formatFuture(diff time.Duration) string {
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := int(diff.Hours() / (24 * 7))
	months := int(diff.Hours() / (24 * 30))
	years := int(diff.Hours() / (24 * 365))

	switch {
	case seconds < 60:
		if seconds <= 1 {
			return "in a moment"
		}
		return fmt.Sprintf("in %d seconds", seconds)
	case minutes < 60:
		if minutes == 1 {
			return "in a minute"
		}
		return fmt.Sprintf("in %d minutes", minutes)
	case hours < 24:
		if hours == 1 {
			return "in an hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	case days < 7:
		if days == 1 {
			return "in a day"
		}
		return fmt.Sprintf("in %d days", days)
	case weeks < 4:
		if weeks == 1 {
			return "in a week"
		}
		return fmt.Sprintf("in %d weeks", weeks)
	case months < 12:
		if months == 1 {
			return "in a month"
		}
		return fmt.Sprintf("in %d months", months)
	default:
		if years == 1 {
			return "in a year"
		}
		return fmt.Sprintf("in %d years", years)
	}
}

// Calendar returns a calendar-style time string
func (m *Moment) Calendar() string {
	now := time.Now()

	// Same day
	if m.time.Year() == now.Year() && m.time.YearDay() == now.YearDay() {
		return "Today at " + m.time.Format("15:04")
	}

	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if m.time.Year() == yesterday.Year() && m.time.YearDay() == yesterday.YearDay() {
		return "Yesterday at " + m.time.Format("15:04")
	}

	// This week
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	if m.time.After(weekStart) {
		return m.time.Format("Monday at 15:04")
	}

	// This year
	if m.time.Year() == now.Year() {
		return m.time.Format("Jan 2 at 15:04")
	}

	// Other years
	return m.time.Format("Jan 2, 2006 at 15:04")
}

// IsToday checks if the time is today
func (m *Moment) IsToday() bool {
	now := time.Now()
	return m.time.Year() == now.Year() && m.time.YearDay() == now.YearDay()
}

// IsYesterday checks if the time is yesterday
func (m *Moment) IsYesterday() bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return m.time.Year() == yesterday.Year() && m.time.YearDay() == yesterday.YearDay()
}

// IsThisWeek checks if the time is this week
func (m *Moment) IsThisWeek() bool {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	return m.time.After(weekStart)
}

// IsThisYear checks if the time is this year
func (m *Moment) IsThisYear() bool {
	now := time.Now()
	return m.time.Year() == now.Year()
}

// Add adds a duration to the time
func (m *Moment) Add(d time.Duration) *Moment {
	return &Moment{time: m.time.Add(d)}
}

// AddDays adds days to the time
func (m *Moment) AddDays(days int) *Moment {
	return &Moment{time: m.time.AddDate(0, 0, days)}
}

// AddMonths adds months to the time
func (m *Moment) AddMonths(months int) *Moment {
	return &Moment{time: m.time.AddDate(0, months, 0)}
}

// AddYears adds years to the time
func (m *Moment) AddYears(years int) *Moment {
	return &Moment{time: m.time.AddDate(years, 0, 0)}
}

// StartOfDay returns the start of the day (00:00:00)
func (m *Moment) StartOfDay() *Moment {
	year, month, day := m.time.Date()
	return &Moment{time: time.Date(year, month, day, 0, 0, 0, 0, m.time.Location())}
}

// EndOfDay returns the end of the day (23:59:59)
func (m *Moment) EndOfDay() *Moment {
	year, month, day := m.time.Date()
	return &Moment{time: time.Date(year, month, day, 23, 59, 59, 999999999, m.time.Location())}
}

// StartOfWeek returns the start of the week (Monday)
func (m *Moment) StartOfWeek() *Moment {
	weekday := int(m.time.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, make it 7
	}
	daysToSubtract := weekday - 1
	return m.AddDays(-daysToSubtract).StartOfDay()
}

// EndOfWeek returns the end of the week (Sunday)
func (m *Moment) EndOfWeek() *Moment {
	weekday := int(m.time.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 0, make it 7
	}
	daysToAdd := 7 - weekday
	return m.AddDays(daysToAdd).EndOfDay()
}

// StartOfMonth returns the start of the month
func (m *Moment) StartOfMonth() *Moment {
	year, month, _ := m.time.Date()
	return &Moment{time: time.Date(year, month, 1, 0, 0, 0, 0, m.time.Location())}
}

// EndOfMonth returns the end of the month
func (m *Moment) EndOfMonth() *Moment {
	year, month, _ := m.time.Date()
	nextMonth := month + 1
	if nextMonth > 12 {
		nextMonth = 1
		year++
	}
	return &Moment{time: time.Date(year, nextMonth, 1, 0, 0, 0, 0, m.time.Location()).Add(-time.Second)}
}

// StartOfYear returns the start of the year
func (m *Moment) StartOfYear() *Moment {
	year, _, _ := m.time.Date()
	return &Moment{time: time.Date(year, 1, 1, 0, 0, 0, 0, m.time.Location())}
}

// EndOfYear returns the end of the year
func (m *Moment) EndOfYear() *Moment {
	year, _, _ := m.time.Date()
	return &Moment{time: time.Date(year+1, 1, 1, 0, 0, 0, 0, m.time.Location()).Add(-time.Second)}
}

// Diff returns the difference between two times
func (m *Moment) Diff(other *Moment) time.Duration {
	return m.time.Sub(other.time)
}

// DiffInDays returns the difference in days
func (m *Moment) DiffInDays(other *Moment) int {
	diff := m.time.Sub(other.time)
	return int(diff.Hours() / 24)
}

// DiffInHours returns the difference in hours
func (m *Moment) DiffInHours(other *Moment) int {
	diff := m.time.Sub(other.time)
	return int(diff.Hours())
}

// DiffInMinutes returns the difference in minutes
func (m *Moment) DiffInMinutes(other *Moment) int {
	diff := m.time.Sub(other.time)
	return int(diff.Minutes())
}

// DiffInSeconds returns the difference in seconds
func (m *Moment) DiffInSeconds(other *Moment) int {
	diff := m.time.Sub(other.time)
	return int(diff.Seconds())
}

// IsBefore checks if the time is before another time
func (m *Moment) IsBefore(other *Moment) bool {
	return m.time.Before(other.time)
}

// IsAfter checks if the time is after another time
func (m *Moment) IsAfter(other *Moment) bool {
	return m.time.After(other.time)
}

// IsSame checks if the time is the same as another time
func (m *Moment) IsSame(other *Moment) bool {
	return m.time.Equal(other.time)
}

// IsSameDay checks if the time is the same day as another time
func (m *Moment) IsSameDay(other *Moment) bool {
	return m.time.Year() == other.time.Year() && m.time.YearDay() == other.time.YearDay()
}

// IsSameMonth checks if the time is the same month as another time
func (m *Moment) IsSameMonth(other *Moment) bool {
	return m.time.Year() == other.time.Year() && m.time.Month() == other.time.Month()
}

// IsSameYear checks if the time is the same year as another time
func (m *Moment) IsSameYear(other *Moment) bool {
	return m.time.Year() == other.time.Year()
}

// String returns a string representation of the time
func (m *Moment) String() string {
	return m.time.String()
}

// Unix returns the Unix timestamp
func (m *Moment) Unix() int64 {
	return m.time.Unix()
}

// UnixMilli returns the Unix timestamp in milliseconds
func (m *Moment) UnixMilli() int64 {
	return m.time.UnixMilli()
}

// UnixNano returns the Unix timestamp in nanoseconds
func (m *Moment) UnixNano() int64 {
	return m.time.UnixNano()
}

// Humanize returns a human-readable string with more context
func (m *Moment) Humanize() string {
	now := time.Now()
	diff := now.Sub(m.time)

	// Handle future times
	if diff < 0 {
		diff = -diff
		return "in " + m.formatDuration(diff)
	}

	return m.formatDuration(diff) + " ago"
}

// formatDuration formats a duration in a human-readable way
func (m *Moment) formatDuration(d time.Duration) string {
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

// RelativeTime returns a relative time string with more precision
func (m *Moment) RelativeTime() string {
	now := time.Now()
	diff := now.Sub(m.time)

	// Handle future times
	if diff < 0 {
		diff = -diff
		return "in " + m.formatPreciseDuration(diff)
	}

	return m.formatPreciseDuration(diff) + " ago"
}

// formatPreciseDuration formats a duration with more precision
func (m *Moment) formatPreciseDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := int(d.Hours() / 24)

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds", seconds)
	case minutes < 60:
		remainingSeconds := seconds % 60
		if remainingSeconds == 0 {
			return fmt.Sprintf("%d minutes", minutes)
		}
		return fmt.Sprintf("%d minutes and %d seconds", minutes, remainingSeconds)
	case hours < 24:
		remainingMinutes := minutes % 60
		if remainingMinutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours and %d minutes", hours, remainingMinutes)
	case days < 7:
		remainingHours := hours % 24
		if remainingHours == 0 {
			return fmt.Sprintf("%d days", days)
		}
		return fmt.Sprintf("%d days and %d hours", days, remainingHours)
	default:
		return fmt.Sprintf("%d days", days)
	}
}
