package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	// Create a temporary directory with a scaffold.yaml
	tmpDir, err := os.MkdirTemp("", "scaffold-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a test scaffold.yaml
	manifestContent := `
name: test-template
description: A test template
type: base
version: "1.0.0"

variables:
  - name: project_name
    description: Name of the project
    required: true
  - name: author
    description: Author name
    default: "Anonymous"
  - name: license
    type: choice
    choices:
      - MIT
      - Apache-2.0
      - GPL-3.0
    default: MIT

files:
  exclude:
    - "*.pyc"
    - "__pycache__"

actions:
  - name: welcome
    type: message
    message: "Project created successfully!"
`
	manifestPath := filepath.Join(tmpDir, "scaffold.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Load the manifest
	manifest, err := LoadManifest(tmpDir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}

	// Verify the manifest
	if manifest.Name != "test-template" {
		t.Errorf("Name = %v, want %v", manifest.Name, "test-template")
	}
	if manifest.Description != "A test template" {
		t.Errorf("Description = %v, want %v", manifest.Description, "A test template")
	}
	if manifest.Type != "base" {
		t.Errorf("Type = %v, want %v", manifest.Type, "base")
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", manifest.Version, "1.0.0")
	}

	// Check variables
	if len(manifest.Variables) != 3 {
		t.Errorf("len(Variables) = %v, want %v", len(manifest.Variables), 3)
	}
	if manifest.Variables[0].Name != "project_name" {
		t.Errorf("Variables[0].Name = %v, want %v", manifest.Variables[0].Name, "project_name")
	}
	if !manifest.Variables[0].Required {
		t.Error("Variables[0].Required should be true")
	}
	if manifest.Variables[1].Default != "Anonymous" {
		t.Errorf("Variables[1].Default = %v, want %v", manifest.Variables[1].Default, "Anonymous")
	}
	if len(manifest.Variables[2].Choices) != 3 {
		t.Errorf("len(Variables[2].Choices) = %v, want %v", len(manifest.Variables[2].Choices), 3)
	}

	// Check files config
	if len(manifest.Files.Exclude) != 2 {
		t.Errorf("len(Files.Exclude) = %v, want %v", len(manifest.Files.Exclude), 2)
	}

	// Check actions
	if len(manifest.Actions) != 1 {
		t.Errorf("len(Actions) = %v, want %v", len(manifest.Actions), 1)
	}
	if manifest.Actions[0].Type != "message" {
		t.Errorf("Actions[0].Type = %v, want %v", manifest.Actions[0].Type, "message")
	}
}

func TestLoadManifest_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scaffold-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = LoadManifest(tmpDir)
	if err == nil {
		t.Error("LoadManifest() should return error for missing manifest")
	}
}

func TestLoadManifest_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scaffold-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write invalid YAML
	manifestPath := filepath.Join(tmpDir, "scaffold.yaml")
	if err := os.WriteFile(manifestPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	_, err = LoadManifest(tmpDir)
	if err == nil {
		t.Error("LoadManifest() should return error for invalid YAML")
	}
}

