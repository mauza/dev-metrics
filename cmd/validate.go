package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mauza/devmetrics/internal"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and repository access",
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	fmt.Println("Validating configuration...")

	config, err := internal.LoadConfig(configFile)
	if err != nil {
		return err
	}

	// Check repositories
	for _, repo := range config.Repositories {
		path := filepath.Clean(repo.Path)
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("Warning: Repository path does not exist: %s\n", path)
		} else {
			fmt.Printf("✓ Repository found: %s\n", path)
		}
	}

	// Check LLM configuration
	modelPath := filepath.Clean(config.LLM.ModelPath)
	if _, err := os.Stat(modelPath); err != nil {
		fmt.Printf("Warning: LLM model not found at %s\n", modelPath)
	} else {
		fmt.Printf("✓ LLM model found: %s\n", modelPath)
	}

	return nil
}
