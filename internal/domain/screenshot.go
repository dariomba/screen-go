package domain

import "time"

type Screenshot struct {
	ID          string
	JobID       string
	StorageKey  string
	ContentType string
	Size        int64
	CreatedAt   time.Time
}
