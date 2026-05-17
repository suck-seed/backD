package backup

import "time"

type Stats struct {
	CopiedFiles  int
	UpdatedFiles int
	SkippedFiles int
	FailedFiles  int
	TotalFiles   int
	CopiedBytes  int64
	CurrentFile  string
	StartedAt    time.Time
	FinishedAt   time.Time
}
