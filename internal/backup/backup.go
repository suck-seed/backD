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

	for _, iSrc := range opts.Sources {
		n, err := countFiles(iSrc)
		if err != nil {

			// Source path missing, we'll record the failure during the walk
			continue
		}

		stats.TotalFiles += n
	}

	if opts.Algorithm == AlgorithmMerkle {
		merkleChanged = make(map[string]bool)

		for _, src := range opts.Sources {
			tree, err := BuildTree(src)
			if err != nil {
				// If we cant build the tree
				// Fall back to copying everything
				allFiles, _ := listAllFiles(src)
				for _, f := range allFiles {

					merkleChanged[f] = true
				}
				continue
			}

			currentRoots = append(currentRoots, tree)

			// Find matching root in previous manifest.
			prevRoot, found := findRootByPath(prevManifest.Roots, src)
			if !found {
				// Entire source is new — mark all its files.
				for _, f := range collectFiles(tree) {
					merkleChanged[f] = true
				}
				continue
			}

			// Only diff paths that actually changed.
			for _, changed := range DiffTrees(tree, prevRoot) {
				merkleChanged[changed] = true
			}
		}
	}

	// Main walk
	for _, src := range opts.Sources {
		walkSource(src, opts, merkleChanged, &stats, ch)
	}

	stats.FinishedAt = time.Now()

	// Merkle Mode save updated manifest
	if opts.Algorithm == AlgorithmMerkle && !opts.DryRun && stats.FailedFiles == 0 {

		newManifest := Manifest{
			TemplateName: opts.TemplateName,
			Algorithm:    string(AlgorithmMerkle),
			CreatedAt: func() time.Time {
				if prevManifest.CreatedAt.IsZero() {
					return stats.StartedAt
				}
				return prevManifest.CreatedAt
			}(),
			UpdatedAt: stats.FinishedAt,
			Roots:     currentRoots,
		}

		// Save this manifest file
		if err := SaveManifest(opts.DeviceMount, opts.TemplateName, newManifest); err != nil {

			ch <- ProgressMsg{
				Stats: stats,
				Done:  true,
				Err:   fmt.Errorf("backup completed but manifest could not be saved: %w", err),
			}
			return
		}

	}

	ch <- ProgressMsg{
		Stats: stats,
		Done:  true,
	}
}

// walkSource processes one top-level source path (file or directory)
func walkSource(src string, opts Options, merkleChanged map[string]bool, stats *Stats, ch chan<- ProgressMsg) {

	srcInfo, err := os.Lstat(src)
	if err != nil {
		stats.FailedFiles++
		sendProgress(ch, *stats)
		return
	}

	if !srcInfo.IsDir() {
		processFile(src, srcInfo, opts, merkleChanged, stats, ch)
		return
	}

	_ = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			stats.FailedFiles++
			sendProgress(ch, *stats)
			return nil // continue walking
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			stats.FailedFiles++
			sendProgress(ch, *stats)
			return nil
		}
		processFile(path, info, opts, merkleChanged, stats, ch)
		return nil
	})
}

// processFile applies the chosen algorithm to a single file and updates stats.
func processFile(src string, srcInfo os.FileInfo, opts Options, merkleChanged map[string]bool, stats *Stats, ch chan<- ProgressMsg) {
	stats.CurrentFile = src
	dst := MirrorSrcToDst(src, opts.DeviceMount)

	switch opts.Algorithm {
	case AlgorithmMetadata:
		processWithMetadata(src, dst, srcInfo, opts.DryRun, stats, ch)

	case AlgorithmHash:
		processWithHash(src, dst, opts.DryRun, stats, ch)

	case AlgorithmMerkle:
		// Merkle mode uses the pre-computed change set for the copy decision,
		// but still delegates the actual copy to the metadata helper so that
		// freshness within a changed subtree is still respected.
		if merkleChanged[src] {
			processWithMetadata(src, dst, srcInfo, opts.DryRun, stats, ch)
		} else {
			stats.SkippedFiles++
			sendProgress(ch, *stats)
		}

	default:
		// Unreachable after validation, but be safe.
		processWithMetadata(src, dst, srcInfo, opts.DryRun, stats, ch)
	}
}

// processWithMetadata applies the metadata comparison strategy.
func processWithMetadata(src, dst string, srcInfo os.FileInfo, dryRum bool, stats *Stats, ch chan<- ProgressMsg) {
	metaDecision := MetadataCheck(srcInfo, dst)

	switch metaDecision {
	case MetadataDecisionSkip:
		stats.SkippedFiles++
		sendProgress(ch, *stats)
		return

	case MetadataDecisionCopy, MetadataDecisionUpdate:

		if dryRum {
			if metaDecision == MetadataDecisionCopy {
				stats.CopiedFiles++
			} else {
				stats.UpdatedFiles++
			}

			sendProgress(ch, *stats)
			return
		}

		existed := metaDecision == MetadataDecisionUpdate
		written, result, err := SafeCopy(src, dst, existed)
		if err != nil {
			stats.FailedFiles++
			sendProgress(ch, *stats)
			return
		}

		applyResult(result, written, stats)
		sendProgress(ch, *stats)
	}

}

// processWithHash applies the SHA-256 content comparision strategy
func processWithHash(src, dst string, dryRun bool, stats *Stats, ch chan<- ProgressMsg) {

	// Fetch hash decision
	hashDecision, err := HashCheck(src, dst)

	// Handle err and HashDecisionFailed
	if err != nil || hashDecision == HashDecisionFailed {
		stats.FailedFiles++
		sendProgress(ch, *stats)
		return
	}

	// Handle other 3 decision
	switch hashDecision {

	case HashDecisionSkip:
		stats.SkippedFiles++
		sendProgress(ch, *stats)

	case HashDecisionCopy, HashDecisionUpdate:

		if dryRun {
			if hashDecision == HashDecisionCopy {
				stats.CopiedFiles++
			} else {
				stats.UpdatedFiles++
			}
			sendProgress(ch, *stats)
			return
		}

		existed := hashDecision == HashDecisionUpdate
		written, result, copyErr := SafeCopy(src, dst, existed)
		if copyErr != nil {
			stats.FailedFiles++
			sendProgress(ch, *stats)
			return
		}

		applyResult(result, written, stats)
		sendProgress(ch, *stats)

	}

}

// applyResult updates stats from a SafeCopy result.
func applyResult(r CopyResult, written int64, stats *Stats) {
	switch r {
	case CopyResultCopied:
		stats.CopiedFiles++
		stats.CopiedBytes += written
	case CopyResultUpdated:
		stats.UpdatedFiles++
		stats.CopiedBytes += written
	case CopyResultFailed:
		stats.FailedFiles++
	}
}

// sendProgress sends a copy of stats on the channel without blocking.
func sendProgress(ch chan<- ProgressMsg, s Stats) {

	select {
	case ch <- ProgressMsg{Stats: s}:
	default:
		// Channel full — drop the intermediate update. The TUI will catch up
		// on the next message. The final Done message is always sent.
	}
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
