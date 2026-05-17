package backup

import "time"

// MerkleNode represents one node in the Merkle tree — either a file (leaf) or
// a directory (internal node).
//
// The JSON tags match the manifest format described in the spec §11.
type MerkleNode struct {
	Path         string       `json:"path"`
	Type         string       `json:"type"` // "file" | "directory"
	Size         int64        `json:"size,omitempty"`
	ModifiedTime time.Time    `json:"modified_time,omitempty"`
	Hash         string       `json:"hash"`
	Children     []MerkleNode `json:"children,omitempty"`
}
