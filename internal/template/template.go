package template

import (
	"backD/internal/config"
	"path/filepath"
	"time"
)

// BackupTemplate represents a saved backup configuration.
type BackupTemplate struct {
	Name      string    `json:"name"`
	Sources   []string  `json:"sources"`
	Algorithm string    `json:"algorithm"`
	CreatedAt time.Time `json:"created_at"`
}

func templatePath(name string) string {
	return filepath.Join(config.TemplateDir(), name+".json")
}
