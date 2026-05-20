package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// HashDecision is the action the hash checker wants to take for a file
type HashDecision int

const (
	// HashDecisionCopy means the destination does not exist — copy it.
	HashDecisionCopy HashDecision = iota

	// HashDecisionUpdate : destination exists but has different content
	HashDecisionUpdate

	// HashDecisionSkip : source and destination are indentical to the byte
	HashDecisionSkip

	// HashDecisionFailed : one of the hash computations failed
	HashDecisionFailed
)

// HashCheck compares src -- destination
// By their SHA-256 content hash
// Returns the action to take and an err if exist
//
// Slower than MetadataCheck as it reads every byte of both file
// But catches rare case where content changed while size and Modification Time did not
func HashCheck(src, dst string) (HashDecision, error) {

	// Source exists, check existance of dst
	if _, err := os.Lstat(dst); err != nil {
		return HashDecisionCopy, nil
	}

	srcHash, err := HashFile(src)
	if err != nil {
		return HashDecisionFailed,
			fmt.Errorf("hash source: %w", err)
	}

	dstHash, err := HashFile(dst)
	if err != nil {
		// Destination exists but cannot read, overwrite it
		return HashDecisionUpdate, nil
	}

	if srcHash != dstHash {
		return HashDecisionUpdate, nil
	}

	return HashDecisionSkip, nil

}

func HashFile(path string) (string, error) {

	// Open path as file
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hash open %s: %w", path, err)
	}

	// close the file on completion
	defer f.Close()

	// hash
	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash read %s: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
