package backup

import "os"

// MetadataDecision is the action the metadata checker wants to take for a file.
type MetadataDecision int

const (
	// MetadataDecisionCopy means the destination does not exist — copy it.
	MetadataDecisionCopy MetadataDecision = iota

	// MetadataDecisionUpdate means the destination exists but is stale — overwrite it.
	MetadataDecisionUpdate

	// MetadataDecisionSkip means the destination is up to date — nothing to do.
	MetadataDecisionSkip
)

// MetadataCheck compares a source file against its destination counterpart
// using size and modification-time metadata and returns the action to take.
//
// Decision logic (matches the spec §6):
//
//  1. If destination does not exist  → Copy
//  2. If source mtime is newer       → Update
//  3. If sizes differ                → Update
//  4. Otherwise                      → Skip
//
// srcInfo must be the result of os.Lstat(src) — the caller already has it
// from the walk, so we avoid a redundant stat.
func MetadataCheck(srcInfo os.FileInfo, dst string) MetadataDecision {

	dtsInfo, err := os.Lstat(dst)
	if err != nil {
		// Destination does not exist
		// Treat it as new
		return MetadataDecisionCopy
	}

	// Compare source info with the dtsInfo
	if srcInfo.ModTime().After(dtsInfo.ModTime()) {
		return MetadataDecisionUpdate
	}

	// Size differes (content changed but ModTime() was not updated
	if srcInfo.Size() != dtsInfo.Size() {
		return MetadataDecisionUpdate
	}

	return MetadataDecisionSkip
}
