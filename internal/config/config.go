package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	GitHub    GitHubConfig `env:"GITHUB_OWNER,required"`
	Sync      SyncConfig   `env:"SYNC_PATH,required"`
	UserAgent string       `env:"USER_AGENT,required"`
}

// GitHubConfig holds the GitHub configuration
type GitHubConfig struct {
	Token         string `env:"GITHUB_TOKEN,required"`
	Owner         string `env:"GITHUB_OWNER,required"`
	RepoName      string `env:"GITHUB_REPO_NAME,required"`
	RepoURL       string `env:"GITHUB_REPO_URL,required"`
	TargetRepoURL string `env:"GITHUB_TARGET_REPO_URL,required"`
	RepoDir       string `env:"GITHUB_REPO_DIR,required"`
	TargetRepoDir string `env:"GITHUB_TARGET_REPO_DIR,required"`
}

// SyncConfig holds synchronization-related configuration
type SyncConfig struct {
	Path            string        `env:"SYNC_PATH,required"`
	OutputDir       string        `env:"SYNC_OUTPUT_DIR,required"`
	PollInterval    string        `env:"SYNC_POLL_INTERVAL,required"`
	PollIntervalDur time.Duration `env:"-"` // Parsed manually
}

// LoadConfig reads environment variables and initializes the Config struct
func LoadConfig() (*Config, error) {
	cfg := Config{}

	// Parse environment variables into the Config struct
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %v", err)
	}

	// Parse PollIntervalDur from PollInterval string
	pollInterval, err := time.ParseDuration(cfg.Sync.PollInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid poll interval duration: %v", err)
	}
	cfg.Sync.PollIntervalDur = pollInterval

	return &cfg, nil
}
