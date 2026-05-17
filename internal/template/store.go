package template

import (
	"backD/internal/config"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func Save(t BackupTemplate) error {

	if err := config.EnsureDirs(); err != nil {
		return err
	}

	// Validate the data
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("Templete name cannot be empty")
	}
	if t.Algorithm == "" {
		t.Algorithm = "metadata"
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	data, err := json.MarshalIndent(t, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(templatePath(t.Name), data, 0644)
}

// Load reads a template by name.
func Load(name string) (BackupTemplate, error) {

	data, err := os.ReadFile(templatePath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return BackupTemplate{}, fmt.Errorf("Template %q not found", name)
		}
		return BackupTemplate{}, err
	}

	var t BackupTemplate
	if err := json.Unmarshal(data, &t); err != nil {
		return BackupTemplate{}, fmt.Errorf("Template %q is invalid JSON: %w", name, err)
	}

	return t, nil
}

func List() (names []string, err error) {

	if err := config.EnsureDirs(); err != nil {
		return nil, err
	}

	// Read Directories
	entries, err := os.ReadDir(config.TemplateDir())
	if err != nil {
		return nil, err
	}

	var n []string

	// loop through the entries
	for _, e := range entries {

		// should be a file and must have suffix json
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {

			n = append(n, strings.TrimSuffix(e.Name(), ".json"))
		}
	}

	return n, nil

}

func Delete(name string) error {

	err := os.Remove(templatePath(name))

	if os.IsNotExist(err) {
		return fmt.Errorf("Template %q not found", name)
	}
	return err
}
