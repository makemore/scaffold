package cmd

import (
	"fmt"
	"sort"

	"github.com/christophercochran/scaffold/internal/registry"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available official and community templates.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	reg := registry.New("")
	templates, err := reg.List()
	if err != nil {
		return fmt.Errorf("failed to load template index: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates available.")
		return nil
	}

	// Sort template names
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Available templates:")
	fmt.Println()

	for _, name := range names {
		entry := templates[name]
		fmt.Printf("  %-12s  %s\n", name, entry.Description)
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  scaffold init myproject --base <template>")
	fmt.Println()
	fmt.Println("Or use a full URL:")
	fmt.Println("  scaffold init myproject --base github:org/repo")

	return nil
}

