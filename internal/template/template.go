package template

import (
	"backD/internal/config"
	"path/filepath"
	"time"
)

// BackupTemplate represents a saved backup configuration.
type BackupTemplate struct {
	Name      string    `yaml:"name"`
	Sources   []string  `yaml:"sources"`
	Algorithm string    `yaml:"algorithm"`
	CreatedAt time.Time `yaml:"created_at"`
}

func templatePath(name string) string {
	return filepath.Join(config.TemplateDir(), name+".yaml")
}
