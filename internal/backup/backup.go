package backup

import (
	"fmt"
	"os"
)

// ProcessMsg is sent on progress channel after every file decision
type ProcessMsg struct {
	Stats Stats
	Done  bool  // true on the final message
	Err   error // non-nil only when the entire run failed before it started
}

// Options control how a backup run behaves
type Options struct {

	// Sources is list of absolute source path (files or directories)
	Sources []string

	// DeviceMount is root mount of the external storage device
	// /run/media/nothing/Sandesh Extra
	DeviceMount string

	// Algorithm selects the comparision strategy
	// Defaults to AlgorithmMetadata
	Algorithm Algorithm

	// TemplateName is used to locate the Merkle manifest. May be empty.
	TemplateName string

	// DryRun, when true, records what *would* happen but does not write any
	// files. Stats will show what would be copied/skipped/updated/failed.
	//? For testing purposes
	DryRun bool
}

// Run performs backup accourding to opts
// Steams ProgressMsg as the returning channel
// <-chan is closed after final ProgressMsg (Done == true) has been sent
func Run(opts Options) <-chan ProcessMsg {

	// Create an channel
	ch := make(chan ProcessMsg, 256)

	// Call run(opts,ch) on another thread
	go func() {
		defer close(ch)
		run(opts, ch)
	}()

	// return the channel back to the calling position
	return ch
}

func run(opts Options, ch chan<- ProcessMsg) {

	if opts.Algorithm == "" {
		opts.Algorithm = AlgorithmMetadata
	}

	// Validate device mount, can be done using os.Stat

	// Stat returns a [FileInfo] describing the named file.
	// If there is an error, it will be of type [*PathError].
	if _, err := os.Stat(opts.DeviceMount); err != nil {
		ch <- ProcessMsg{
			Err:  fmt.Errorf("Device mount %s not accessible: %w", opts.DeviceMount, err),
			Done: true,
		}
		return
	}

	// Make a new stats with the starting time as now
	// stats := Stats{StartedAt: time.Now()}
}
