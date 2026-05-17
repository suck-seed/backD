package backup

import "time"

// BackDFolder is the top-level folder created on the external device.
const BackDFolder = "backD"

// backdMetaDir is the hidden metadata directory inside BackDFolder.
const backdMetaDir = ".backd"

// Manifest stores the Merkle tree state of a previous backup run so that
// Merkle Mode can quickly determine which folders have changed.
//
// One manifest file is kept per template, stored at:
//
//	<deviceMount>/backD/.backd/templates/<templateName>-manifest.json
//
// A global manifest (not tied to a template) is stored at:
//
//	<deviceMount>/backD/.backd/manifest.json
type Manifest struct {
	TemplateName string       `json:"template_name"`
	Algorithm    string       `json:"algorithm"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Roots        []MerkleNode `json:"roots"`
}
