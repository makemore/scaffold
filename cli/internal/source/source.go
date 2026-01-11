// Package source handles fetching templates from various sources
package source

import (
	"fmt"
	"net/url"
	"strings"
)

// Type represents the type of template source
type Type string

const (
	TypeGit   Type = "git"
	TypeFile  Type = "file"
	TypeURL   Type = "url"
)

// Source represents a parsed template source
type Source struct {
	Type     Type
	URI      string   // Original URI
	URL      string   // Resolved URL/path
	Ref      string   // Git ref (tag, branch, commit)
	Subdir   string   // Subdirectory within the source
	Provider string   // For git: github, gitlab, bitbucket, etc.
}

// Parse parses a source URI string into a Source struct
// Supported formats:
//   - git:https://github.com/org/repo
//   - git:git@github.com:org/repo.git
//   - git:https://github.com/org/repo#v1.0
//   - git:https://github.com/org/repo//subdir#main
//   - file:./relative/path
//   - file:~/templates/my-template
//   - file:/absolute/path
//   - https://example.com/template.tar.gz
//   - github:org/repo
//   - gitlab:org/repo
//   - bitbucket:org/repo
func Parse(uri string) (*Source, error) {
	if uri == "" {
		return nil, fmt.Errorf("empty source URI")
	}

	// Handle shorthand aliases
	if strings.HasPrefix(uri, "github:") {
		path := strings.TrimPrefix(uri, "github:")
		return parseGitSource("https://github.com/" + path)
	}
	if strings.HasPrefix(uri, "gitlab:") {
		path := strings.TrimPrefix(uri, "gitlab:")
		return parseGitSource("https://gitlab.com/" + path)
	}
	if strings.HasPrefix(uri, "bitbucket:") {
		path := strings.TrimPrefix(uri, "bitbucket:")
		return parseGitSource("https://bitbucket.org/" + path)
	}

	// Handle explicit prefixes
	if strings.HasPrefix(uri, "git:") {
		return parseGitSource(strings.TrimPrefix(uri, "git:"))
	}
	if strings.HasPrefix(uri, "file:") {
		return parseFileSource(strings.TrimPrefix(uri, "file:"))
	}

	// Handle plain URLs
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return parseURLSource(uri)
	}

	return nil, fmt.Errorf("unknown source format: %s", uri)
}

func parseGitSource(uri string) (*Source, error) {
	s := &Source{
		Type: TypeGit,
		URI:  uri,
	}

	// Extract ref (after #)
	if idx := strings.LastIndex(uri, "#"); idx != -1 {
		s.Ref = uri[idx+1:]
		uri = uri[:idx]
	}

	// Extract subdir (after //)
	if idx := strings.Index(uri, "//"); idx != -1 {
		s.Subdir = uri[idx+2:]
		uri = uri[:idx]
	}

	s.URL = uri

	// Detect provider
	if strings.Contains(uri, "github.com") {
		s.Provider = "github"
	} else if strings.Contains(uri, "gitlab.com") {
		s.Provider = "gitlab"
	} else if strings.Contains(uri, "bitbucket.org") {
		s.Provider = "bitbucket"
	}

	return s, nil
}

func parseFileSource(path string) (*Source, error) {
	return &Source{
		Type: TypeFile,
		URI:  path,
		URL:  path,
	}, nil
}

func parseURLSource(uri string) (*Source, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	return &Source{
		Type: TypeURL,
		URI:  uri,
		URL:  parsed.String(),
	}, nil
}

// String returns a human-readable representation of the source
func (s *Source) String() string {
	result := fmt.Sprintf("%s:%s", s.Type, s.URL)
	if s.Subdir != "" {
		result += "//" + s.Subdir
	}
	if s.Ref != "" {
		result += "#" + s.Ref
	}
	return result
}

