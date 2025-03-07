package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mauza/devmetrics/internal"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
)

var (
	repoPath string
	days     int
	persona  string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate git commits with realistic changes",
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&repoPath, "repo-path", "", "Path to repository (overrides config file)")
	generateCmd.Flags().IntVar(&days, "days", 7, "Number of days to generate commits for")
	generateCmd.Flags().StringVar(&persona, "persona", "", "Developer persona to use (early_bird, night_owl, balanced)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load config
	config, err := internal.LoadConfig(configFile)
	if err != nil {
		return err
	}

	// Get API key from environment variable
	apiKey := os.Getenv(config.LLM.APIKeyEnvVar)
	if apiKey == "" {
		return fmt.Errorf("API key environment variable %s is not set", config.LLM.APIKeyEnvVar)
	}

	// Initialize components
	llm, err := internal.NewLLMOperations(
		config.LLM.Provider,
		config.LLM.Endpoint,
		apiKey,
		config.LLM.Model,
		config.LLM.Temperature,
		config.LLM.MaxTokens,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM: %w", err)
	}
	defer llm.Close()

	patternGen := internal.NewCommitPatternGenerator()

	// Generate commit patterns
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	patterns := patternGen.GeneratePatterns(startDate, endDate, persona)

	// Determine repositories to process
	var repositories []internal.Repository

	if repoPath != "" {
		repositories = []internal.Repository{
			{Path: repoPath, Patterns: []string{"*.*"}},
		}
	} else {
		repositories = config.Repositories
	}

	// Process each repository
	for _, repo := range repositories {
		gitOps, err := internal.NewGitOperations(repo.Path)
		if err != nil {
			fmt.Printf("Skipping repository due to error: %v\n", err)
			continue
		}

		// Verify repository access
		if err := gitOps.VerifyRepoAccess(); err != nil {
			fmt.Printf("Skipping repository due to access issues: %v\n", err)
			continue
		}

		fmt.Printf("Generating commits for %s\n", repo.Path)

		// Process each commit pattern
		for _, pattern := range patterns {
			// Get list of files we can modify
			modifiableFiles, err := gitOps.GetModifiableFiles(repo.Patterns)
			if err != nil {
				fmt.Printf("Error getting modifiable files: %v\n", err)
				continue
			}

			if len(modifiableFiles) == 0 {
				fmt.Printf("No matching files found in %s\n", repo.Path)
				continue
			}

			// Select files to modify
			numFiles := min(pattern.NumFiles, len(modifiableFiles))
			filesToModify := selectRandomFiles(modifiableFiles, numFiles)

			var changesDescription []string

			// Modify each file
			for _, filePath := range filesToModify {
				content, err := gitOps.ReadFile(filePath)
				if err != nil {
					continue
				}

				// Generate changes using LLM
				newContent, changeDesc, err := llm.GenerateCodeChanges(filePath, content)
				if err != nil {
					continue
				}

				// Apply changes
				if err := gitOps.ModifyFile(filePath, newContent); err != nil {
					continue
				}

				changesDescription = append(changesDescription,
					fmt.Sprintf("%s: %s", filepath.Base(filePath), changeDesc))
			}

			if len(changesDescription) > 0 {
				// Generate commit message
				changesSummary := fmt.Sprintf("%s\n\nChanges:\n%s",
					pattern.Description,
					formatChanges(changesDescription))

				commitMsg, err := llm.GenerateCommitMessage(changesSummary)
				if err != nil {
					fmt.Printf("Error generating commit message: %v\n", err)
					continue
				}

				// Create commit with pattern timestamp
				if err := gitOps.CreateCommit(commitMsg, filesToModify, &pattern.Timestamp); err != nil {
					fmt.Printf("Failed to create commit: %v\n", err)
				} else {
					fmt.Printf("Created commit: %s\n", commitMsg)
				}
			}
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func selectRandomFiles(files []string, n int) []string {
	// Fisher-Yates shuffle and take first n elements
	result := make([]string, len(files))
	copy(result, files)

	for i := len(result) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result[:n]
}

func formatChanges(changes []string) string {
	var result string
	for _, change := range changes {
		result += fmt.Sprintf("- %s\n", change)
	}
	return result
}
