package config

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "backD")
}

func TemplateDir() string {
	return filepath.Join(ConfigDir(), "template")
}

func LogDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "backD")
}

func LogFile() string {
	return filepath.Join(LogDir(), "backD.log")
}

func EnsureDirs() error {
	dirs := []string{ConfigDir(), TemplateDir(), LogDir()}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
	}

	return nil
}
