package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	info, err := os.Lstat(path)
	if err != nil {
		return MerkleNode{}, fmt.Errorf("lstat %s: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return MerkleNode{}, fmt.Errorf("symlink %s skipped", path)
	}

	if !info.IsDir() {
		return buildFileNode(path, info)
	}
	return buildDirNode(path)
}

func buildFileNode(path string, info os.FileInfo) (MerkleNode, error) {
	h, err := HashFile(path)
	if err != nil {
		return MerkleNode{}, err
	}
	return MerkleNode{
		Path:         path,
		Type:         "file",
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		Hash:         h,
	}, nil
}

func buildDirNode(path string) (MerkleNode, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return MerkleNode{}, fmt.Errorf("readdir %s: %w", path, err)
	}

	// Sort by name for stable hashing.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var children []MerkleNode
	for _, e := range entries {
		child, err := BuildTree(filepath.Join(path, e.Name()))
		if err != nil {
			// Symlinks and unreadable entries are silently skipped so a single
			// bad file doesn't abort the entire tree.
			continue
		}
		children = append(children, child)
	}

	h := hashDirChildren(children)
	return MerkleNode{
		Path:     path,
		Type:     "directory",
		Hash:     h,
		Children: children,
	}, nil
}

// hashDirChildren computes the folder hash from its children as described in
// the spec §10.2:
//
//	SHA256(child1_name + child1_hash + child2_name + child2_hash …)
//
// Children must already be sorted by name before this is called.
func hashDirChildren(children []MerkleNode) string {
	h := sha256.New()
	for _, c := range children {
		h.Write([]byte(filepath.Base(c.Path)))
		h.Write([]byte(c.Hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// FindNode searches a tree for the node at the given absolute path.
// Returns (node, true) if found, or (MerkleNode{}, false) if not.
func FindNode(root MerkleNode, path string) (MerkleNode, bool) {
	if root.Path == path {
		return root, true
	}
	for _, c := range root.Children {
		if n, ok := FindNode(c, path); ok {
			return n, true
		}
	}
	return MerkleNode{}, false
}

// DiffTrees compares a current tree against a previous tree and returns the
// set of source paths that need to be copied.
//
// The algorithm matches the spec §12:
//   - If a root's hash is unchanged → skip the entire subtree.
//   - If a directory hash changed → recurse into children.
//   - If a file hash changed or is new → include it.
func DiffTrees(current, previous MerkleNode) []string {
	// Unchanged subtree — skip everything beneath it.
	if current.Hash == previous.Hash {
		return nil
	}

	if !current.IsDir() {
		// File changed.
		return []string{current.Path}
	}

	// Directory changed — check children.
	var changed []string
	for _, curChild := range current.Children {
		prevChild, found := FindNode(previous, curChild.Path)
		if !found {
			// New entry — collect all files beneath it.
			changed = append(changed, collectFiles(curChild)...)
			continue
		}
		changed = append(changed, DiffTrees(curChild, prevChild)...)
	}
	return changed
}

// collectFiles returns the path of every file leaf in a subtree.
func collectFiles(node MerkleNode) []string {
	if !node.IsDir() {
		return []string{node.Path}
	}
	var paths []string
	for _, c := range node.Children {
		paths = append(paths, collectFiles(c)...)
	}
	return paths
}
