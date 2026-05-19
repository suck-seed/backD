package backup

import (
	"fmt"
	"time"
)

// Algorithm represents the backup comparison mode.
type Algorithm string

const (
	AlgorithmMetadata Algorithm = "metadata"
	AlgorithmHash     Algorithm = "hash"
	AlgorithmMerkle   Algorithm = "merkle"
)

func ParseAlgorithm(s string) (Algorithm, error) {
	switch Algorithm(s) {
	case AlgorithmHash, AlgorithmMerkle, AlgorithmMetadata:
		return Algorithm(s), nil

	default:
		return "", fmt.Errorf("Unknown backup algorithm %q", s)
	}
}

// Stats tracks counters and timing for a single backup run.
//
// The backup engine sends a *copy* of Stats on every progress message so the
// TUI goroutine can read it safely without synchronisation.
type Stats struct {
	// File counters
	CopiedFiles  int // brand-new files written to the destination
	UpdatedFiles int // files that already existed but were newer/different
	SkippedFiles int // files identical to destination — no copy needed
	FailedFiles  int // files that could not be read or written

	// CopiedBytes is the total number of bytes written this run (copies + updates).
	CopiedBytes int64

	// TotalFiles is set before the walk starts so the progress bar is accurate.
	TotalFiles int

	// CurrentFile is the absolute source path being processed right now.
	CurrentFile string

	// Timing
	StartedAt  time.Time
	FinishedAt time.Time // zero while the backup is still running
}

func (s Stats) ProgressPercent() int {
	if s.TotalFiles == 0 {
		return 0
	}

	// Proessed()
	p := (s.Processed() / s.TotalFiles) * 100
	if p > 100 {
		return 100
	}

	return p
}

// Processed returns the number of files that have recieved a final decision
// Decision can be copied, updated, skipped, failed
func (s Stats) Processed() int {

	return s.CopiedFiles + s.UpdatedFiles + s.SkippedFiles + s.FailedFiles

}

// Elapsed returns a human-readable duration since StartedAt.
// If the backup has finished it returns the exact run time.
func (s Stats) Elapsed() string {
	end := s.FinishedAt

	if end.IsZero() {
		end = time.Now()
	}

	dur := end.Sub(s.StartedAt).Round(time.Second)

	hour := int(dur.Hours())
	min := int(dur.Minutes()) % 60
	sec := int(dur.Seconds()) % 60

	switch {
	case hour > 0:
		return fmt.Sprintf("%dh %dm %ds", hour, min, sec)

	case min > 0:
		return fmt.Sprintf("%dm %ds", min, sec)

	default:
		return fmt.Sprintf("%ds", sec)

	}

}

func FormatBytes(b uint64) string {
	const unit uint64 = 1024
	if b < uint64(unit) {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
