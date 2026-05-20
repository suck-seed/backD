package backup

import (
	"time"
)

// MerkleNode represents one node in the Merkle tree — either a file (leaf) or
// a directory (internal node).
type MerkleNode struct {
	Path         string       `yaml:"path"`
	Type         string       `yaml:"type"` // "file" | "directory"
	Size         int64        `yaml:"size,omitempty"`
	ModifiedTime time.Time    `yaml:"modified_time,omitempty"`
	Hash         string       `yaml:"hash"`
	Children     []MerkleNode `yaml:"children,omitempty"`
}

// IsDir returns true when the node represents a directory
func (n MerkleNode) IsDir() bool {
	return n.Type == "directory"
}

func BuildTree(path string) (MerkleNode, error) {

}
