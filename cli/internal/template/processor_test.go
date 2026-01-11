package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makemore/scaffold/internal/config"
)

func TestProcessor_Process(t *testing.T) {
	// Create source directory with template files
	srcDir, err := os.MkdirTemp("", "scaffold-src")
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// Create destination directory
	destDir, err := os.MkdirTemp("", "scaffold-dest")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Create test files in source
	testFiles := map[string]string{
		"README.md":                      "# {{ project_name }}\n\nBy {{ author }}",
		"__project_slug__/config.py":     "PROJECT = '{{ project_slug }}'",
		"__project_slug__/settings.yaml": "name: {{ project_name }}",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(srcDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Create scaffold.yaml
	manifestContent := `
name: test-template
type: base
variables:
  - name: project_name
  - name: project_slug
  - name: author
`
	if err := os.WriteFile(filepath.Join(srcDir, "scaffold.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Load manifest and create processor
	manifest, err := config.LoadManifest(srcDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	processor := NewProcessor(manifest, srcDir, destDir)
	processor.SetVariables(map[string]string{
		"project_name": "My Project",
		"project_slug": "my_project",
		"author":       "Test Author",
	})

	// Process the template
	if err := processor.Process(); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// Verify README.md
	readmeContent, err := os.ReadFile(filepath.Join(destDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}
	if !strings.Contains(string(readmeContent), "# My Project") {
		t.Errorf("README.md should contain '# My Project', got: %s", readmeContent)
	}
	if !strings.Contains(string(readmeContent), "By Test Author") {
		t.Errorf("README.md should contain 'By Test Author', got: %s", readmeContent)
	}

	// Verify directory was renamed
	configPath := filepath.Join(destDir, "my_project", "config.py")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("my_project/config.py should exist (directory renamed)")
	}

	// Verify content substitution in renamed directory
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.py: %v", err)
	}
	if !strings.Contains(string(configContent), "PROJECT = 'my_project'") {
		t.Errorf("config.py should contain \"PROJECT = 'my_project'\", got: %s", configContent)
	}
}

func TestProcessor_SkipsScaffoldYaml(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "scaffold-src")
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	destDir, err := os.MkdirTemp("", "scaffold-dest")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	// Create scaffold.yaml
	if err := os.WriteFile(filepath.Join(srcDir, "scaffold.yaml"), []byte("name: test"), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest := &config.Manifest{Name: "test"}
	processor := NewProcessor(manifest, srcDir, destDir)

	if err := processor.Process(); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	// scaffold.yaml should NOT be copied
	if _, err := os.Stat(filepath.Join(destDir, "scaffold.yaml")); !os.IsNotExist(err) {
		t.Error("scaffold.yaml should not be copied to destination")
	}
}

