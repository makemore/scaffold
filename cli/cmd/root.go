package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"
	// Commit is set at build time
	Commit = "none"
)

var rootCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Bootstrap any software stack with sensible defaults",
	Long: `Scaffold is a cross-platform CLI that composes modular templates
into production-ready project scaffolds.

Sensible defaults and best practices baked in—great for humans,
even better for AI-assisted development.

Templates can be sourced from:
  • Any git repository (GitHub, GitLab, Bitbucket, self-hosted)
  • Local file paths
  • URLs (tarballs, zip archives)

Example:
  scaffold init myapp --base git:https://github.com/org/template
  scaffold init myapp --base file:~/templates/base --add file:./modules/postgres`,
	Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

