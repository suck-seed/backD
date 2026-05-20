package backup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProgressMsg is sent on progress channel after every file decision
type ProgressMsg struct {
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
func Run(opts Options) <-chan ProgressMsg {

	// Create an channel
	ch := make(chan ProgressMsg, 256)

	// Call run(opts,ch) on another thread
	go func() {
		defer close(ch)
		run(opts, ch)
	}()

	// return the channel back to the calling position
	return ch
}

func run(opts Options, ch chan<- ProgressMsg) {

	if opts.Algorithm == "" {
		opts.Algorithm = AlgorithmMetadata
	}

	// Validate device mount, can be done using os.Stat

	// Stat returns a [FileInfo] describing the named file.
	// If there is an error, it will be of type [*PathError].
	if _, err := os.Stat(opts.DeviceMount); err != nil {
		ch <- ProgressMsg{
			Err:  fmt.Errorf("Device mount %s not accessible: %w", opts.DeviceMount, err),
			Done: true,
		}
		return
	}

	// Make a new stats with the starting time as now
	stats := Stats{StartedAt: time.Now()}

	// For merkle mode : loading previous manifest and building current backup tree
	var (
		prevManifest  Manifest
		currentRoots  []MerkleNode
		merkleChanged map[string]bool // Tracking changed/updated files/folders compaared to prevoius Manifest
	)

	if opts.Algorithm == AlgorithmMerkle {
		prevManifest, _ = LoadManifest(opts.DeviceMount, opts.TemplateName)
		// No need to handle error
		// Missing a manifest will -> copy everything and make a new manifest
	}

	// for _, iSrc := range opts.Sources {

	// }
}

// countFiles returns the number of regular files reachable from path.
func countFiles(path string) (int, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 1, nil
	}
	count := 0
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count, nil
}

// listAllFiles returns all regular file paths beneath root.
func listAllFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

// findRootByPath finds a MerkleNode in a slice whose Path matches path.
func findRootByPath(roots []MerkleNode, path string) (MerkleNode, bool) {
	for _, r := range roots {
		if r.Path == path {
			return r, true
		}
	}
	return MerkleNode{}, false
}

// Validate checks Options for common mistakes before Run is called.
func (o Options) Validate() error {
	if o.DeviceMount == "" {
		return ErrNoDevice
	}
	if len(o.Sources) == 0 {
		return ErrNoSources
	}
	if o.Algorithm != "" {
		if _, err := ParseAlgorithm(string(o.Algorithm)); err != nil {
			return err
		}
	}
	return nil
}

// --- Sentinel errors used by callers ---

// ErrNoSources is returned when Options.Sources is empty.
var ErrNoSources = errors.New("no source paths specified")

// ErrNoDevice is returned when Options.DeviceMount is empty.
var ErrNoDevice = errors.New("no backup device specified")
