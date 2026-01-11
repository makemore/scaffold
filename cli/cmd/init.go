package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/makemore/scaffold/internal/config"
	"github.com/makemore/scaffold/internal/registry"
	"github.com/makemore/scaffold/internal/source"
	"github.com/makemore/scaffold/internal/template"
	"github.com/spf13/cobra"
)

var (
	baseTemplate string
	addModules   []string
	variables    []string
	outputDir    string
	noPrompt     bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project from templates",
	Long: `Initialize a new project by composing a base template with optional modules.

Templates can be specified using various source formats:

Git repositories:
  git:https://github.com/org/repo
  git:git@github.com:org/repo.git
  git:https://gitlab.com/org/repo#v1.0
  git:https://github.com/org/repo//subdir#main

Local file paths:
  file:./relative/path
  file:~/templates/my-template
  file:/absolute/path

URLs:
  https://example.com/template.tar.gz
  https://example.com/template.zip

Shorthand aliases:
  github:org/repo
  gitlab:org/repo
  bitbucket:org/repo`,
	Example: `  # Interactive mode
  scaffold init myapp

  # With base template
  scaffold init myapp --base github:org/base-template

  # With modules
  scaffold init myapp \
    --base git:https://github.com/org/template \
    --add file:~/modules/postgres \
    --add github:org/mod-auth

  # With variables (non-interactive)
  scaffold init myapp \
    --base github:org/template \
    --var project_name=myapp \
    --var org=myorg \
    --no-prompt`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&baseTemplate, "base", "b", "", "Base template source")
	initCmd.Flags().StringArrayVarP(&addModules, "add", "a", nil, "Additional modules to layer")
	initCmd.Flags().StringArrayVarP(&variables, "var", "v", nil, "Variables in key=value format")
	initCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (defaults to project name)")
	initCmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "Disable interactive prompts")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	// If no project name and interactive mode, prompt for it
	if projectName == "" && !noPrompt {
		projectName = prompt("Project name")
		if projectName == "" {
			return fmt.Errorf("project name is required")
		}
	}

	if projectName == "" {
		return fmt.Errorf("project name is required (or use interactive mode)")
	}

	// Determine output directory
	outDir := outputDir
	if outDir == "" {
		outDir = projectName
	}

	// Check if output directory already exists
	if _, err := os.Stat(outDir); err == nil {
		return fmt.Errorf("directory %s already exists", outDir)
	}

	// If no base template specified, prompt or show list
	if baseTemplate == "" && !noPrompt {
		reg := registry.New("")
		templates, _ := reg.List()

		fmt.Println("\nAvailable templates:")
		for name, entry := range templates {
			fmt.Printf("  %s - %s\n", name, entry.Description)
		}
		fmt.Println()

		baseTemplate = prompt("Template (or full URL)")
		if baseTemplate == "" {
			return fmt.Errorf("template is required")
		}
	}

	if baseTemplate == "" {
		return fmt.Errorf("--base template is required (or use interactive mode)")
	}

	fmt.Printf("🚀 Creating project: %s\n", projectName)

	// Resolve template shorthand to full source
	reg := registry.New("")
	resolvedSource, err := reg.Resolve(baseTemplate)
	if err != nil {
		return fmt.Errorf("failed to resolve template: %w", err)
	}

	fmt.Printf("📦 Template: %s\n", resolvedSource)

	// Parse the source URI
	src, err := source.Parse(resolvedSource)
	if err != nil {
		return fmt.Errorf("failed to parse source: %w", err)
	}

	// Fetch the template
	fmt.Println("⬇️  Fetching template...")
	fetcher := source.NewFetcher("")
	templatePath, err := fetcher.Fetch(src)
	if err != nil {
		return fmt.Errorf("failed to fetch template: %w", err)
	}

	// Load the manifest
	manifest, err := config.LoadManifest(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Collect variables
	vars := collectVariables(manifest, projectName)

	// Prompt for missing required variables
	if !noPrompt {
		for _, v := range manifest.Variables {
			if _, ok := vars[v.Name]; !ok {
				defaultVal := v.Default
				promptText := v.Name
				if v.Description != "" {
					promptText = v.Description
				}
				if defaultVal != "" {
					promptText += fmt.Sprintf(" [%s]", defaultVal)
				}

				val := prompt(promptText)
				if val == "" {
					val = defaultVal
				}
				if val == "" && v.Required {
					return fmt.Errorf("variable %s is required", v.Name)
				}
				vars[v.Name] = val
			}
		}
	}

	// Create output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process template
	fmt.Println("📝 Processing template...")
	processor := template.NewProcessor(manifest, templatePath, outDir)
	processor.SetVariables(vars)

	if err := processor.Process(); err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	fmt.Printf("\n✅ Project created at: %s\n", outDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", outDir)

	// Show post-generation actions
	for _, action := range manifest.Actions {
		if action.Type == "message" {
			fmt.Printf("  %s\n", action.Message)
		}
	}

	return nil
}

func collectVariables(manifest *config.Manifest, projectName string) map[string]string {
	vars := make(map[string]string)

	// Set project_name and common variants
	vars["project_name"] = projectName
	vars["project_slug"] = strings.ReplaceAll(strings.ToLower(projectName), "-", "_")

	// Parse --var flags
	for _, v := range variables {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	// Apply defaults for missing variables
	for _, v := range manifest.Variables {
		if _, ok := vars[v.Name]; !ok && v.Default != "" {
			vars[v.Name] = v.Default
		}
	}

	return vars
}

func prompt(label string) string {
	fmt.Printf("%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// absPath returns the absolute path, handling ~ expansion
func absPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	abs, _ := filepath.Abs(path)
	return abs
}

