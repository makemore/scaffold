package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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
		prompt := &survey.Input{Message: "Project name:"}
		if err := survey.AskOne(prompt, &projectName, survey.WithValidator(survey.Required)); err != nil {
			return err
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

		// Build options list
		options := make([]string, 0, len(templates)+1)
		names := make([]string, 0, len(templates))
		for name := range templates {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			entry := templates[name]
			options = append(options, fmt.Sprintf("%s - %s", name, entry.Description))
		}
		options = append(options, "Other (enter URL)")

		var selection string
		prompt := &survey.Select{
			Message: "Select a template:",
			Options: options,
		}
		if err := survey.AskOne(prompt, &selection); err != nil {
			return err
		}

		if selection == "Other (enter URL)" {
			urlPrompt := &survey.Input{Message: "Template URL:"}
			if err := survey.AskOne(urlPrompt, &baseTemplate, survey.WithValidator(survey.Required)); err != nil {
				return err
			}
		} else {
			// Extract template name from selection
			baseTemplate = strings.SplitN(selection, " - ", 2)[0]
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
				message := v.Name
				if v.Description != "" {
					message = v.Description
				}

				var val string
				var err error

				switch v.Type {
				case "select", "choice":
					if len(v.Choices) > 0 {
						prompt := &survey.Select{
							Message: message,
							Options: v.Choices,
							Default: v.Default,
						}
						err = survey.AskOne(prompt, &val)
					} else {
						prompt := &survey.Input{Message: message, Default: v.Default}
						err = survey.AskOne(prompt, &val)
					}
				case "confirm", "boolean":
					var confirm bool
					prompt := &survey.Confirm{
						Message: message,
						Default: v.Default == "true",
					}
					err = survey.AskOne(prompt, &confirm)
					if confirm {
						val = "true"
					} else {
						val = "false"
					}
				default:
					prompt := &survey.Input{Message: message, Default: v.Default}
					if v.Required {
						err = survey.AskOne(prompt, &val, survey.WithValidator(survey.Required))
					} else {
						err = survey.AskOne(prompt, &val)
					}
				}

				if err != nil {
					return err
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

	// Process additional modules
	for _, moduleSource := range addModules {
		fmt.Printf("📦 Adding module: %s\n", moduleSource)

		// Resolve module source
		resolvedModule, err := reg.Resolve(moduleSource)
		if err != nil {
			return fmt.Errorf("failed to resolve module %s: %w", moduleSource, err)
		}

		// Parse the module source
		moduleSrc, err := source.Parse(resolvedModule)
		if err != nil {
			return fmt.Errorf("failed to parse module source: %w", err)
		}

		// Fetch the module
		modulePath, err := fetcher.Fetch(moduleSrc)
		if err != nil {
			return fmt.Errorf("failed to fetch module: %w", err)
		}

		// Load module manifest
		moduleManifest, err := config.LoadManifest(modulePath)
		if err != nil {
			return fmt.Errorf("failed to load module manifest: %w", err)
		}

		// Prompt for module-specific variables
		if !noPrompt {
			for _, v := range moduleManifest.Variables {
				if _, ok := vars[v.Name]; !ok {
					message := v.Name
					if v.Description != "" {
						message = v.Description
					}

					var val string
					prompt := &survey.Input{Message: message, Default: v.Default}
					if err := survey.AskOne(prompt, &val); err != nil {
						return err
					}
					vars[v.Name] = val
				}
			}
		}

		// Process module (layer on top of existing files)
		moduleProcessor := template.NewProcessor(moduleManifest, modulePath, outDir)
		moduleProcessor.SetVariables(vars)

		if err := moduleProcessor.Process(); err != nil {
			return fmt.Errorf("failed to process module %s: %w", moduleSource, err)
		}

		// Collect module actions
		manifest.Actions = append(manifest.Actions, moduleManifest.Actions...)
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

// absPath returns the absolute path, handling ~ expansion
func absPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	abs, _ := filepath.Abs(path)
	return abs
}

