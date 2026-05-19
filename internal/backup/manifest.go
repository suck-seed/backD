package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.yaml.in/yaml/v3"
)

// BackDFolder is the top-level folder created on the external device.
const BackDFolder = "backD"

// backdMetaDir is the hidden metadata directory inside BackDFolder.
const backdMetaDir = ".backd"

// Manifest stores the Merkle tree state of a previous backup run so that
// Merkle Mode can quickly determine which folders have changed.
//
// One manifest file is kept per template, stored at:
//
//	<deviceMount>/backD/.backd/templates/<templateName>-manifest.yaml
//
// A global manifest (not tied to a template) is stored at:
//
//	<deviceMount>/backD/.backd/manifest.yaml
type Manifest struct {
	TemplateName string       `yaml:"template_name"`
	Algorithm    string       `yaml:"algorithm"`
	CreatedAt    time.Time    `yaml:"created_at"`
	UpdatedAt    time.Time    `yaml:"updated_at"`
	Roots        []MerkleNode `yaml:"roots"`
}

// ManifestPath returns the path for a template-specific manifest on a device
// If templateName == ", fallbacks to the global ManifestPath
func ManifestPath(deviceMount, templateName string) string {

	base := filepath.Join(deviceMount, backdMetaDir, backdMetaDir)
	if templateName == "" {
		return filepath.Join(base, "manifest.yaml")
	}

	return filepath.Join(base, "templates", templateName+"-manifest.yaml")
}

func LoadManifest(deviceMount, templateName string) (Manifest, error) {

	path := ManifestPath(deviceMount, templateName)

	// Read the file
	f, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Manifest{}, fmt.Errorf("Manifest %s not found", templateName)
		}
		return Manifest{}, fmt.Errorf("Failed Reading %s", templateName)
	}

	// Unmarshall and send it back
	var m Manifest

	if err := yaml.Unmarshal(f, &m); err != nil {
		return Manifest{}, fmt.Errorf("Failed Unmarshalling : %s : %v ", path, err)
	}

	// Success, return Manifest
	return m, nil

}
