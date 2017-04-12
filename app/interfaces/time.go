package interfaces

import (
	"time"
)

// Time provides the current time.
type Time interface {
	// Now returns the current time in UTC.
	Now() time.Time
}
