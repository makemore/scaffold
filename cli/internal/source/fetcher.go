package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Fetcher handles fetching templates from various sources
type Fetcher struct {
	CacheDir string
}

// NewFetcher creates a new Fetcher with the given cache directory
func NewFetcher(cacheDir string) *Fetcher {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".scaffold", "cache")
	}
	return &Fetcher{CacheDir: cacheDir}
}

// Fetch retrieves a template from the given source and returns the local path
func (f *Fetcher) Fetch(src *Source) (string, error) {
	switch src.Type {
	case TypeGit:
		return f.fetchGit(src)
	case TypeFile:
		return f.fetchFile(src)
	case TypeURL:
		return f.fetchURL(src)
	default:
		return "", fmt.Errorf("unsupported source type: %s", src.Type)
	}
}

func (f *Fetcher) fetchGit(src *Source) (string, error) {
	// Create a unique cache path based on the URL
	cachePath := f.cachePathFor(src)

	// Check if already cached
	if _, err := os.Stat(cachePath); err == nil {
		// TODO: Check if we need to update (fetch latest)
		return f.resolveSubdir(cachePath, src.Subdir), nil
	}

	// Clone the repository
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	args := []string{"clone", "--depth", "1"}
	if src.Ref != "" {
		args = append(args, "--branch", src.Ref)
	}
	args = append(args, src.URL, cachePath)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return f.resolveSubdir(cachePath, src.Subdir), nil
}

func (f *Fetcher) fetchFile(src *Source) (string, error) {
	path := src.URL

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Make relative paths absolute
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	// Verify the path exists
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("template path does not exist: %s", path)
	}

	return path, nil
}

func (f *Fetcher) fetchURL(src *Source) (string, error) {
	// TODO: Implement URL fetching (download and extract archives)
	return "", fmt.Errorf("URL fetching not yet implemented")
}

func (f *Fetcher) cachePathFor(src *Source) string {
	// Create a safe directory name from the URL
	safeName := strings.ReplaceAll(src.URL, "/", "_")
	safeName = strings.ReplaceAll(safeName, ":", "_")
	safeName = strings.ReplaceAll(safeName, "@", "_")

	if src.Ref != "" {
		safeName += "_" + src.Ref
	}

	return filepath.Join(f.CacheDir, safeName)
}

func (f *Fetcher) resolveSubdir(basePath, subdir string) string {
	if subdir == "" {
		return basePath
	}
	return filepath.Join(basePath, subdir)
}

