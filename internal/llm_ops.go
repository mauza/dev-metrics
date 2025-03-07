package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/mauza/gollm"
)

// LLMOperations handles interactions with the LLM model
type LLMOperations struct {
	llm gollm.LLM
}

// NewLLMOperations creates a new LLM operations instance
func NewLLMOperations(provider string, endpoint string, apiKey string, model string, temperature float64, maxTokens int) (*LLMOperations, error) {
	llm, err := gollm.NewLLM(
		gollm.SetProvider(provider),
		gollm.SetModel(model),
		gollm.SetEndpoint(endpoint),
		gollm.SetAPIKey(apiKey),
		gollm.SetMaxTokens(maxTokens),
		gollm.SetTemperature(temperature),
		gollm.SetMaxRetries(3),
		gollm.SetTimeout(300*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM model: %w", err)
	}

	return &LLMOperations{
		llm: llm,
	}, nil
}

// GenerateCommitMessage generates a commit message based on the changes
func (l *LLMOperations) GenerateCommitMessage(changes string) (string, error) {
	prompt := gollm.NewPrompt(fmt.Sprintf(`Given these code changes:

%s

Generate a concise, professional git commit message following these rules:
- Use present tense
- Start with a verb
- Be specific but concise
- Max 72 characters for first line
- Optional: Add detailed description after blank line`, changes),
		gollm.WithDirectives(
			"Be professional",
			"Be specific",
			"Use conventional commit format",
		),
	)

	response, err := l.llm.Generate(context.Background(), prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return response, nil
}

// GenerateCodeChanges generates changes for a given file
func (l *LLMOperations) GenerateCodeChanges(filePath, content string) (string, string, error) {
	prompt := gollm.NewPrompt(fmt.Sprintf(`Review and suggest improvements for this code:

File: %s

%s

Provide ONLY the final version of the code with minimal, realistic improvements.`, filePath, content),
		gollm.WithDirectives(
			"Make minimal necessary changes",
			"Maintain code style",
			"Focus on readability and maintainability",
		),
		gollm.WithOutput("Respond with only the improved code"),
	)

	response, err := l.llm.Generate(context.Background(), prompt)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate code changes: %w", err)
	}

	// Create a brief description of changes
	descPrompt := gollm.NewPrompt(fmt.Sprintf("Summarize the changes made to %s in one brief sentence", filePath))
	description, err := l.llm.Generate(context.Background(), descPrompt)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate change description: %w", err)
	}

	return response, description, nil
}

// Close releases any resources
func (l *LLMOperations) Close() error {
	return nil
}
