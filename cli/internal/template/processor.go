// Package template handles template processing and variable substitution
package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/christophercochran/scaffold/internal/config"
)

// Processor handles template processing
type Processor struct {
	manifest  *config.Manifest
	variables map[string]string
	srcDir    string
	destDir   string
}

// NewProcessor creates a new template processor
func NewProcessor(manifest *config.Manifest, srcDir, destDir string) *Processor {
	return &Processor{
		manifest:  manifest,
		variables: make(map[string]string),
		srcDir:    srcDir,
		destDir:   destDir,
	}
}

// SetVariables sets the template variables
func (p *Processor) SetVariables(vars map[string]string) {
	p.variables = vars
}

// Process processes the template and writes to the destination
func (p *Processor) Process() error {
	return filepath.Walk(p.srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(p.srcDir, path)
		if err != nil {
			return err
		}

		// Skip root directory marker
		if relPath == "." {
			return nil
		}

		// Skip scaffold.yaml
		if relPath == "scaffold.yaml" {
			return nil
		}

		// Skip hidden files/directories (but not the root)
		baseName := filepath.Base(relPath)
		if strings.HasPrefix(baseName, ".") && baseName != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply variable substitution to path
		destRelPath := p.substituteInPath(relPath)
		destPath := filepath.Join(p.destDir, destRelPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return p.processFile(path, destPath, info.Mode())
	})
}

func (p *Processor) processFile(srcPath, destPath string, mode os.FileMode) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Check if file is binary
	if isBinary(srcPath) {
		return copyFile(srcPath, destPath, mode)
	}

	// Read and process text file
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	processed := p.substituteVariables(string(content))

	return os.WriteFile(destPath, []byte(processed), mode)
}

// substituteVariables replaces {{ variable }} patterns
func (p *Processor) substituteVariables(content string) string {
	// Match {{ variable_name }} with optional whitespace
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name
		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		varName := submatch[1]

		if val, ok := p.variables[varName]; ok {
			return val
		}
		return match // Keep original if not found
	})
}

// substituteInPath handles __variable__ patterns in file/directory names
func (p *Processor) substituteInPath(path string) string {
	// Match __variable_name__ pattern
	re := regexp.MustCompile(`__([a-zA-Z_][a-zA-Z0-9_]*)__`)

	return re.ReplaceAllStringFunc(path, func(match string) string {
		varName := strings.Trim(match, "_")
		if val, ok := p.variables[varName]; ok {
			return val
		}
		return match
	})
}

func isBinary(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

