package backup

import (
	"errors"
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

// SaveManifest serialises m and mount it automatically to the corrrect location
// on the external device
// Creates any missing directories
func SaveManifest(deviceMount, templateName string, m Manifest) error {

	path := ManifestPath(deviceMount, templateName)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	// Write to a tmp file, then rename
	// For atomicity and to prevent potential race condition
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write manifest tmp: %w", err)
	}

	// Rename
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename manifest %s -> %s: %w", tmp, path, err)
	}

	return nil

}

func DeleteManifest(deviceMount, templateName string) error {
	path := ManifestPath(deviceMount, templateName)
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("delete manifest path %s: %w", path, err)
	}

	return nil
}

func ManifestExists(deviceMount, templateName string) bool {

	_, err := os.Lstat(ManifestPath(deviceMount, templateName))

	return err == nil
}

// Sentinel errors for manifest operations.
var (
	// ErrManifestNotFound is returned by LoadManifest when no manifest exists
	// this is normal on the first backup run.
	ErrManifestNotFound = errors.New("manifest not found")

	// ErrManifestCorrupted is returned when a manifest file cannot be decoded.
	ErrManifestCorrupted = errors.New("manifest corrupted")
)
