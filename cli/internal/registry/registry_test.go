package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_Resolve(t *testing.T) {
	// Create a temporary index file
	tmpDir, err := os.MkdirTemp("", "scaffold-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	indexContent := `
version: "1"
official:
  django:
    source: "github:makemore/scaffold//templates/django-base"
    description: "Django template"
  nextjs:
    source: "github:makemore/scaffold//templates/nextjs-base"
    description: "Next.js template"
aliases:
  django-api: django
  next: nextjs
`
	indexPath := filepath.Join(tmpDir, "templates.yaml")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}

	// Set environment variable to use our test index
	os.Setenv("SCAFFOLD_INDEX", indexPath)
	defer os.Unsetenv("SCAFFOLD_INDEX")

	reg := New(tmpDir)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "resolve official template",
			input: "django",
			want:  "github:makemore/scaffold//templates/django-base",
		},
		{
			name:  "resolve alias",
			input: "django-api",
			want:  "github:makemore/scaffold//templates/django-base",
		},
		{
			name:  "resolve nextjs alias",
			input: "next",
			want:  "github:makemore/scaffold//templates/nextjs-base",
		},
		{
			name:  "passthrough full URL",
			input: "github:other/repo",
			want:  "github:other/repo",
		},
		{
			name:  "passthrough git URL",
			input: "git:https://example.com/repo",
			want:  "git:https://example.com/repo",
		},
		{
			name:  "passthrough file path",
			input: "file:./local/path",
			want:  "file:./local/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reg.Resolve(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scaffold-registry-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	indexContent := `
version: "1"
official:
  django:
    source: "github:makemore/scaffold//templates/django-base"
    description: "Django REST API"
  nextjs:
    source: "github:makemore/scaffold//templates/nextjs-base"
    description: "Next.js with TypeScript"
`
	indexPath := filepath.Join(tmpDir, "templates.yaml")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}

	os.Setenv("SCAFFOLD_INDEX", indexPath)
	defer os.Unsetenv("SCAFFOLD_INDEX")

	reg := New(tmpDir)
	templates, err := reg.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("List() returned %d templates, want 2", len(templates))
	}

	if templates["django"].Description != "Django REST API" {
		t.Errorf("django description = %v, want 'Django REST API'", templates["django"].Description)
	}

	if templates["nextjs"].Description != "Next.js with TypeScript" {
		t.Errorf("nextjs description = %v, want 'Next.js with TypeScript'", templates["nextjs"].Description)
	}
}

