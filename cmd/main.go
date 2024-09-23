package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"io/fs"
	"strings"

	"github.com/morpho-app/lexsync/internal/config"
	"github.com/morpho-app/lexsync/internal/converter"
	"github.com/morpho-app/lexsync/internal/github"
	"github.com/morpho-app/lexsync/internal/logger"
	"github.com/morpho-app/lexsync/internal/utils"
	"go.uber.org/zap"
)

func main() {
	zapLogger := logger.GetLogger()
	defer logger.SyncLogger()

	sugar := zapLogger.Sugar()

	cfg, err := config.LoadConfig()
	if err != nil {
		sugar.Fatalf("Error loading config: %v", err)
	}

	sugar.Infof("Configuration Initialized: %+v", cfg)

	// Initialize GitHub client with the default BaseURL
	githubClient := github.NewGitHubClient(
		cfg.GitHub.Token,
		cfg.UserAgent,
		cfg.GitHub.Owner,
		cfg.GitHub.RepoName,
		"", // Empty BaseURL defaults to "https://api.github.com"
	)

	var lastCommitSha string

	// Start an infinite loop to periodically check for updates
	for {
		// Fetch the latest commit SHA from the GitHub repository
		latestCommitSha, err := githubClient.GetLatestCommitSha(cfg.Sync.Path, sugar)
		if err != nil {
			sugar.Errorf("Error fetching latest commit SHA: %v", err)
			// Wait for an hour before retrying in case of failure
			time.Sleep(1 * time.Hour)
			continue
		}

		// Compare the latest commit SHA with the last processed one
		if latestCommitSha != lastCommitSha {
			sugar.Info("Detected changes in the lexicons directory")

			// Pull the latest changes from the repository
			if err := PullRepo(githubClient, cfg, sugar); err != nil {
				sugar.Errorf("Error pulling repository: %v", err)
				// Wait for 10 minutes before retrying
				time.Sleep(10 * time.Minute)
				continue
			}

			// Convert JSON lexicons to Kotlin data classes
			if err := ConvertLexicons(cfg, sugar); err != nil {
				sugar.Errorf("Error converting lexicons: %v", err)
				// Wait for 10 minutes before retrying
				time.Sleep(10 * time.Minute)
				continue
			}

			// Commit and push the changes to the target repository
			if err := CommitAndPush(githubClient, cfg, sugar); err != nil {
				sugar.Errorf("Error pushing to repository: %v", err)
				// Wait for 10 minutes before retrying
				time.Sleep(10 * time.Minute)
				continue
			}

			// Update the last processed commit SHA
			lastCommitSha = latestCommitSha
		} else {
			sugar.Info("No new changes detected in the Lexicons directory")
		}

		// Wait for the specified poll interval before the next check
		sugar.Infof("Sleeping for %s before the next check...", cfg.Sync.PollIntervalDur)
		time.Sleep(cfg.Sync.PollIntervalDur)
	}
}

// PullRepo clones the repository if it doesn't exist locally or pulls the latest changes.
func PullRepo(client *github.GitHubClient, cfg *config.Config, sugar *zap.SugaredLogger) error {
	// Clone the repository if the local directory doesn't exist
	if _, err := os.Stat(cfg.GitHub.RepoDir); os.IsNotExist(err) {
		sugar.Infof("Cloning repository from %s to %s...", cfg.GitHub.RepoURL, cfg.GitHub.RepoDir)
		if err := client.CloneRepo(cfg.GitHub.RepoURL, cfg.GitHub.RepoDir, sugar); err != nil {
			return fmt.Errorf("failed to clone repository: %v", err)
		}
		sugar.Info("Repository cloned successfully.")
	} else {
		sugar.Infof("Repository directory %s exists. Pulling latest changes...", cfg.GitHub.RepoDir)
		// Pull the latest changes
		if err := client.PullLatest(cfg.GitHub.RepoDir, sugar); err != nil {
			return fmt.Errorf("failed to pull latest changes: %v", err)
		}
		sugar.Info("Repository updated successfully.")
	}
	return nil
}

// ConvertLexicons traverses all JSON files in the specified lexicons directory and converts each to a corresponding Kotlin file.
func ConvertLexicons(cfg *config.Config, sugar *zap.SugaredLogger) error {
	lexiconsDir := filepath.Join(cfg.GitHub.RepoDir, cfg.Sync.Path)
	outputDir := cfg.Sync.OutputDir

	sugar.Infof("Accessing lexicons directory: %s", lexiconsDir)

	// Verify that the lexicons directory exists
	if _, err := os.Stat(lexiconsDir); os.IsNotExist(err) {
		return fmt.Errorf("lexicons directory does not exist at path: %s", lexiconsDir)
	}

	// Create the output directory if it doesn't exist
	if err := utils.EnsureDir(outputDir); err != nil {
		return err
	}

	// Initialize a WaitGroup to handle concurrency
	var wg sync.WaitGroup

	// Use WalkDir to traverse the lexicons directory recursively
	err := filepath.WalkDir(lexiconsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			sugar.Warnf("Error accessing path %s: %v", path, err)
			return nil // Continue walking despite the error
		}

		if d.IsDir() {
			return nil // Skip directories
		}

		// Process only JSON files
		if filepath.Ext(d.Name()) == ".json" {
			wg.Add(1)
			go func(filePath string) {
				defer wg.Done()

				// Determine the relative path to maintain directory structure in the output
				relPath, err := filepath.Rel(lexiconsDir, filePath)
				if err != nil {
					sugar.Warnf("Error determining relative path for %s: %v", filePath, err)
					return
				}

				// Determine the corresponding Kotlin file path
				kotlinFileName := strings.TrimSuffix(filepath.Base(relPath), ".json") + ".kt"
				kotlinFileDir := filepath.Join(outputDir, filepath.Dir(relPath))
				kotlinFilePath := filepath.Join(kotlinFileDir, kotlinFileName)

				// Ensure the destination directory exists
				if err := utils.EnsureDir(kotlinFileDir); err != nil {
					sugar.Warnf("Error creating directory %s: %v", kotlinFileDir, err)
					return
				}

				// Process the JSON file to generate Kotlin code
				kotlinCode, err := converter.ProcessLexiconJSON(filePath, sugar)
				if err != nil {
					sugar.Warnf("Error processing %s: %v", filePath, err)
					return
				}

				// Write the generated Kotlin code to the destination file
				if err := os.WriteFile(kotlinFilePath, []byte(kotlinCode), 0644); err != nil {
					sugar.Warnf("Error writing Kotlin file %s: %v", kotlinFilePath, err)
					return
				}

				sugar.Infof("Successfully converted %s to %s", filePath, kotlinFilePath)
			}(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path %s: %v", lexiconsDir, err)
	}

	// Wait for all goroutines to finish processing
	wg.Wait()

	sugar.Infof("All JSON files have been converted to Kotlin data classes in %s", outputDir)
	return nil
}

// CommitAndPush stages all changes, commits them with a predefined message, and pushes to the target repository.
func CommitAndPush(client *github.GitHubClient, cfg *config.Config, sugar *zap.SugaredLogger) error {
	commitMessage := "Syncing new lexicon files"
	if err := client.CommitAndPush(cfg.GitHub.TargetRepoDir, commitMessage, sugar); err != nil {
		return fmt.Errorf("failed to commit and push changes: %v", err)
	}
	sugar.Info("Changes committed and pushed successfully.")
	return nil
}
