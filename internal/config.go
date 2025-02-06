package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM struct {
		MaxTokens   int     `yaml:"max_tokens"`
		ModelPath   string  `yaml:"model_path"`
		Temperature float64 `yaml:"temperature"`
	} `yaml:"llm"`
	Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
	Path     string   `yaml:"path"`
	Patterns []string `yaml:"patterns"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if len(config.Repositories) == 0 {
		return nil, fmt.Errorf("no repositories configured in config.yaml")
	}
	if config.LLM.ModelPath == "" {
		return nil, fmt.Errorf("no LLM configuration in config.yaml")
	}

	return &config, nil
}

func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
