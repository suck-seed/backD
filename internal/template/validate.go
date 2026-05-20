package template

import "os"

// Validate checks whether all source paths still exist.
func Validate(name string) ([]string, error) {
	t, err := Load(name)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, src := range t.Sources {
		if _, err := os.Stat(src); os.IsNotExist(err) {
			missing = append(missing, src)
		}
	}
	return missing, nil
}
