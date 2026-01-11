package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ManifestFile = "scaffold.yaml"
	LockFile     = "scaffold.lock"
)

// LoadManifest loads a scaffold.yaml from the given directory
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s found in %s", ManifestFile, dir)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// SaveLockfile writes a lockfile to the given directory
func SaveLockfile(dir string, lock *Lockfile) error {
	path := filepath.Join(dir, LockFile)

	data, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lockfile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	return nil
}

// LoadLockfile loads a scaffold.lock from the given directory
func LoadLockfile(dir string) (*Lockfile, error) {
	path := filepath.Join(dir, LockFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No lockfile is not an error
		}
		return nil, fmt.Errorf("failed to read lockfile: %w", err)
	}

	var lock Lockfile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse lockfile: %w", err)
	}

	return &lock, nil
}

