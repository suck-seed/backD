package backup

import "time"

// MerkleNode represents one node in the Merkle tree — either a file (leaf) or
// a directory (internal node).
//
// The YAML tags match the manifest format described in the spec §11.
type MerkleNode struct {
	Path         string       `yaml:"path"`
	Type         string       `yaml:"type"` // "file" | "directory"
	Size         int64        `yaml:"size,omitempty"`
	ModifiedTime time.Time    `yaml:"modified_time,omitempty"`
	Hash         string       `yaml:"hash"`
	Children     []MerkleNode `yaml:"children,omitempty"`
}
