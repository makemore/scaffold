// Package registry handles the template index for shorthand names
package registry

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed templates.yaml
var embeddedIndex embed.FS

const (
	// RemoteIndexURL is the URL to fetch the latest template index
	RemoteIndexURL = "https://raw.githubusercontent.com/scaffold-dev/scaffold/main/templates.yaml"
	// CacheExpiry is how long to cache the remote index
	CacheExpiry = 24 * time.Hour
)

// Index represents the templates.yaml structure
type Index struct {
	Version   string                     `yaml:"version"`
	Official  map[string]TemplateEntry   `yaml:"official"`
	Community map[string]TemplateEntry   `yaml:"community"`
	Aliases   map[string]string          `yaml:"aliases"`
}

// TemplateEntry represents a single template in the index
type TemplateEntry struct {
	Source      string `yaml:"source"`
	Description string `yaml:"description"`
}

// Registry manages template lookups
type Registry struct {
	index    *Index
	cacheDir string
}

// New creates a new Registry
func New(cacheDir string) *Registry {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".scaffold", "cache")
	}
	return &Registry{cacheDir: cacheDir}
}

// Resolve looks up a shorthand name and returns the full source URI
// Returns the original name if not found (allows pass-through of full URIs)
func (r *Registry) Resolve(name string) (string, error) {
	if err := r.ensureLoaded(); err != nil {
		return name, nil // Fall back to treating as URI
	}

	// Check aliases first
	if aliasTarget, ok := r.index.Aliases[name]; ok {
		name = aliasTarget
	}

	// Check official templates
	if entry, ok := r.index.Official[name]; ok {
		return entry.Source, nil
	}

	// Check community templates
	if entry, ok := r.index.Community[name]; ok {
		return entry.Source, nil
	}

	// Not found - return original (might be a full URI)
	return name, nil
}

// List returns all available templates
func (r *Registry) List() (map[string]TemplateEntry, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}

	result := make(map[string]TemplateEntry)
	for k, v := range r.index.Official {
		result[k] = v
	}
	for k, v := range r.index.Community {
		result[k] = v
	}
	return result, nil
}

func (r *Registry) ensureLoaded() error {
	if r.index != nil {
		return nil
	}

	// Check for local index override (for development)
	if localPath := os.Getenv("SCAFFOLD_INDEX"); localPath != "" {
		if idx, err := r.loadFromFile(localPath); err == nil {
			r.index = idx
			return nil
		}
	}

	// Try to load from cache first
	if idx, err := r.loadFromCache(); err == nil {
		r.index = idx
		return nil
	}

	// Try to fetch from remote
	if idx, err := r.fetchRemote(); err == nil {
		r.index = idx
		_ = r.saveToCache(idx)
		return nil
	}

	// Fall back to embedded index
	return r.loadEmbedded()
}

func (r *Registry) loadFromFile(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func (r *Registry) loadFromCache() (*Index, error) {
	cachePath := filepath.Join(r.cacheDir, "templates.yaml")
	info, err := os.Stat(cachePath)
	if err != nil {
		return nil, err
	}

	// Check if cache is expired
	if time.Since(info.ModTime()) > CacheExpiry {
		return nil, fmt.Errorf("cache expired")
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func (r *Registry) fetchRemote() (*Index, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(RemoteIndexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func (r *Registry) saveToCache(idx *Index) error {
	if err := os.MkdirAll(r.cacheDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(idx)
	if err != nil {
		return err
	}

	cachePath := filepath.Join(r.cacheDir, "templates.yaml")
	return os.WriteFile(cachePath, data, 0644)
}

func (r *Registry) loadEmbedded() error {
	data, err := embeddedIndex.ReadFile("templates.yaml")
	if err != nil {
		return fmt.Errorf("failed to load embedded index: %w", err)
	}

	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return fmt.Errorf("failed to parse embedded index: %w", err)
	}

	r.index = &idx
	return nil
}

