package dataset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

// SaveManifest writes the dataset manifest to the specified path atomically using a temp file.
func SaveManifest(path string, manifest *domain.DatasetManifest) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("write temp manifest: %w", err)
	}

	if err := os.Rename(tempFile, path); err != nil {
		return fmt.Errorf("rename manifest: %w", err)
	}

	return nil
}

// LoadManifest reads and parses the dataset manifest from the given path.
func LoadManifest(path string) (*domain.DatasetManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest domain.DatasetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}

	return &manifest, nil
}
