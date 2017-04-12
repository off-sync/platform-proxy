package time

import "time"

// SystemTime implements the Time interface using the system time
// as its source.
type SystemTime struct{}

// NewSystemTime returns a new system time.
func NewSystemTime() *SystemTime {
	return &SystemTime{}
}

// Now returns the current system time in UTC.
func (t *SystemTime) Now() time.Time {
	return time.Now().UTC()
}
