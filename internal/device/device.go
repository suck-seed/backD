package device

import (
	"os"
	"path/filepath"
	"strings"
)

// ExternalDevice represents a mounted external storage device.
type ExternalDevice struct {
	Name       string
	MountPath  string
	FileSystem string

	// Total Space
	// uint64 for accomodating high byte count
	// and prevent negative
	TotalSpace uint64

	// Available Space
	AvailableSpace uint64

	// Free Space
	FreeSpace uint64
}

func Detect() ([]*ExternalDevice, error) {

	var devices []*ExternalDevice

	home, _ := os.UserHomeDir()
	username := filepath.Base(home)

	// Searchable file paths in linux
	searchPath := []string{
		"/mnt",
		filepath.Join("/run/media", username),
		filepath.Join("/media", username),
	}

	// Registry of the filepath seen
	// pathname -> bool
	seen := map[string]bool{}

	// loop through the searchPath
	for _, base := range searchPath {

		// get entries
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}

		// Found entriessss, looping through
		for _, e := range entries {

			if !e.IsDir() {
				continue
			}
			// If e is directory, mount
			mount := filepath.Join(base, e.Name())

			if seen[mount] {
				// already seen
				continue
			}

			// Check if the external storage device is writeable or not
			if !isWritable(mount) {
				continue
			}

			fsType, err := fsTypeForMount(mount)
			if err != nil {
				continue
			}

			total, avail, free, err := bytesFromStatfs(mount)
			if err != nil {
				continue
			}

			devices = append(devices, &ExternalDevice{
				Name:           formatName(e.Name()),
				MountPath:      mount,
				FileSystem:     fsType,
				TotalSpace:     total,
				AvailableSpace: avail,
				FreeSpace:      free,
			})

		}

	}

	return devices, nil
}

// formatName turns a mount label into a human-readable name.
func formatName(label string) string {
	// Replace underscores and dashes with spaces, title-case
	name := strings.ReplaceAll(label, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return name
}
