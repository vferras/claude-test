package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey      string
	DatabaseURL string
	Symbols     []string
}

type stocksFile struct {
	Symbols []string `yaml:"symbols"`
}

func Load(yamlPath string) (*Config, error) {
	apiKey := os.Getenv("MARKETSTACK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MARKETSTACK_API_KEY env var is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL env var is required")
	}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading stocks config: %w", err)
	}

	var sf stocksFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parsing stocks config: %w", err)
	}

	if len(sf.Symbols) == 0 {
		return nil, fmt.Errorf("no symbols configured in %s", yamlPath)
	}

	return &Config{
		APIKey:      apiKey,
		DatabaseURL: dbURL,
		Symbols:     sf.Symbols,
	}, nil
}
