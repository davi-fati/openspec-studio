package domain

import "time"

// Project is a registered OpenSpec project directory tracked by the Studio.
type Project struct {
	ID           int64
	Name         string
	Path         string
	LastOpenedAt time.Time
	Available    bool
}
